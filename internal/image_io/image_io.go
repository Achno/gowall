package imageio

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/utils"
)

// We could use a library for this purpose

func isPng(magicBytes []byte) bool {
	png := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	for index, currentByte := range png {
		if magicBytes[index] != currentByte {
			return false
		}
	}
	return true
}

func isJpeg(magicBytes []byte) bool {
	jpeg := []byte{0xFF, 0xD8, 0xFF}
	for index, currentByte := range jpeg {
		if magicBytes[index] != currentByte {
			return false
		}
	}
	return true
}

func isWebp(magicBytes []byte) bool {
	webp := []byte{0x52, 0x49, 0x46, 0x46}
	for index, currentByte := range webp {
		if magicBytes[index] != currentByte {
			return false
		}
	}
	return true
}

func readImageFormat(imgSrc ImageReader) (string, error) {
	magicBytes := make([]byte, 12)

	reader, err := imgSrc.Open()
	if err != nil {
		return "", err
	}

	if _, err := io.ReadFull(reader, magicBytes); err != nil {
		return "", err
	}
	if isPng(magicBytes) {
		return "png", nil
	}
	if isWebp(magicBytes) {
		return "webp", nil
	}
	if isJpeg(magicBytes) {
		return "jpeg", nil
	}
	return "", fmt.Errorf("uknown format")
}

// Filters out all files other than .png, .jpeg, .jpg, .webp in a directory
func filterImages(entries []fs.DirEntry) ([]string, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("directory is empty")
	}

	var imageFiles []string

	supportedExtensions := map[string]bool{
		".png":  true,
		".jpeg": true,
		".jpg":  true,
		".webp": true,
	}

	for _, entry := range entries {
		if !entry.IsDir() && supportedExtensions[strings.ToLower(filepath.Ext(entry.Name()))] {
			imageFiles = append(imageFiles, entry.Name())
		}
	}

	return imageFiles, nil
}

func GetImagesFromDirectoryRecursively(path string) ([]FileReader, error) {
	var files []FileReader
	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !config.SupportedExtensions[ext] {
			files = append(files, FileReader{Path: path})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no image files found in directory or subdirectories")
	}

	return files, nil
}

type ImageIO struct {
	ImageInput  ImageReader
	ImageOutput ImageWriter
	Format      string
	Theme       string
}

// Input image abstraction
type ImageReader interface {
	Open() (*os.File, error)
	String() string
}

// Ouput image abstraction
type ImageWriter interface {
	Create() (*os.File, error)
	String() string
}

type FileReader struct {
	Path string
}

type FileWriter struct {
	Path string
}

type (
	Stdin  struct{}
	Stdout struct{}
)

func (fr FileReader) Open() (*os.File, error) {
	f, err := os.Open(fr.Path)
	f.Close()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs FileReader) String() string {
	filePath, _ := filepath.Abs(fs.Path)
	return filePath
}

func (fw FileWriter) Create() (*os.File, error) {
	f, err := os.Create(fw.Path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fw FileWriter) String() string {
	filePath, _ := filepath.Abs(fw.Path)
	return filePath
}

func (ss Stdin) Open() (*os.File, error) {
	return os.Stdin, nil
}

func (ss Stdin) String() string {
	return "/dev/stdin"
}

func (so Stdout) Create() (*os.File, error) {
	return os.Stdout, nil
}

func (so Stdout) String() string {
	return "/dev/stdout"
}

// DetermineImageOperations generates ImageIO structs based on program flags and command io arguments.
func DetermineImageOperations(flags config.GlobalSubCommandFlags, args []string) []ImageIO {
	// Process by priority: directory > batch files > single file/stdin
	if flags.InputDir != "" {
		return processDirectoryImages(flags)
	}

	if len(flags.InputFiles) > 0 {
		return processBatchFiles(flags)
	}

	return processSingleFile(flags, args)
}

// processSingleFile handles both file and STDIN input cases
func processSingleFile(flags config.GlobalSubCommandFlags, args []string) []ImageIO {
	input := determineInput(args)
	output, ext := determineOutput(flags, args, input)

	return []ImageIO{
		{
			ImageInput:  input,
			ImageOutput: output,
			Format:      ext,
			Theme:       flags.Theme,
		},
	}
}

// determineInput resolves the input source (file or stdin)
func determineInput(args []string) ImageReader {
	// If no args or first arg is "-", use stdin
	if len(args) == 0 || args[0] == "-" {
		return Stdin{}
	}

	// Otherwise use file
	return FileReader{Path: args[0]}
}

// determineOutput resolves the output destination and format
func determineOutput(flags config.GlobalSubCommandFlags, args []string, input ImageReader) (ImageWriter, string) {
	// Check if output should be stdout
	if IsStdoutOutput(flags, args) {
		ext, err := determineFileExt(flags, input, Stdout{})
		utils.HandleError(err, "could not determine file extension")
		return Stdout{}, ext
	}

	outputDest := resolveOutputPath(flags, args, input)

	// Create appropriate writer and get format
	output := FileWriter{Path: outputDest}
	ext, err := determineFileExt(flags, input, output)
	utils.HandleError(err, "could not determine file extension")

	return output, ext
}

// resolveOutputPath determines the final output path based on flags and args
func resolveOutputPath(flags config.GlobalSubCommandFlags, args []string, input ImageReader) string {
	var outputDest string

	// Priority for output destination:
	// 1. --output flag
	// 2. Second positional argument
	// 3. Generated path based on input
	if flags.OutputDestination != "" {
		outputDest = flags.OutputDestination
	} else if len(args) > 1 {
		outputDest = args[1]
	} else {
		outputDest = generateOutputPath(flags, args, input)
	}

	// Resolve absolute path and ensure correct extension
	absPath, err := filepath.Abs(outputDest)
	utils.HandleError(err, "could not resolve absolute path for output destination")

	// Determine final directory
	var finalDir string
	if flags.OutputDestination != "" && shouldRenameFile(flags.OutputDestination) {
		finalDir = config.GowallConfig.OutputFolder
	} else {
		finalDir = filepath.Dir(absPath)
	}

	ext, err := determineFileExt(flags, input, FileWriter{Path: outputDest})
	utils.HandleError(err, "could not determine file extension")

	return filepath.Join(finalDir, replaceExt(filepath.Base(absPath), ext))
}

// generateOutputPath creates an output path when none is specified
func generateOutputPath(flags config.GlobalSubCommandFlags, args []string, input ImageReader) string {
	// For stdin input, generate timestamp-based filename
	if len(args) == 0 || args[0] == "-" {
		ts := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("img-%s", ts)
		ext, err := determineFileExt(flags, input, nil)
		utils.HandleError(err, "could not determine file extension")
		return filepath.Join(config.GowallConfig.OutputFolder, filename+"."+ext)
	}

	// For file input, base output on input filename
	absInput, err := filepath.Abs(args[0])
	utils.HandleError(err, "could not resolve absolute path for input")
	baseName := filepath.Base(absInput)
	ext, err := determineFileExt(flags, input, nil)
	utils.HandleError(err, "could not determine file extension")
	return filepath.Join(config.GowallConfig.OutputFolder, replaceExt(baseName, ext))
}

// IsStdoutOutput checks if the output destination indicates stdout
func IsStdoutOutput(flags config.GlobalSubCommandFlags, args []string) bool {
	return flags.OutputDestination == "-" ||
		flags.OutputDestination == "/dev/stdout" ||
		(len(args) > 1 && args[1] == "-")
}

// shouldRenameFile determines if the input file should be renamed
func shouldRenameFile(outputDest string) bool {
	isAbs := filepath.IsAbs(outputDest)
	if !isAbs {
		return false
	}
	if filepath.Ext(outputDest) != "" {
		return false
	}
	return true
}

// Determines file extension based on the output flags, format flags and
// extensions from the io arguments
func determineFileExt(flags config.GlobalSubCommandFlags, input ImageReader, output ImageWriter) (string, error) {
	// If there is format flag return format
	if flags.Format != "" {
		return flags.Format, nil
	}

	// If there is an output flag and has ext return that ext
	if ext := filepath.Ext(flags.OutputDestination); ext != "" {
		return ext, nil
	}

	// Check if output is a FileWriter to get its path
	if fileWriter, ok := output.(FileWriter); ok {
		if ext := filepath.Ext(fileWriter.Path); ext != "" {
			return ext, nil
		}
	}

	// Check if input is a FileReader to get its path
	if fileReader, ok := input.(FileReader); ok {
		if ext := filepath.Ext(fileReader.Path); ext != "" {
			return ext, nil
		}
	}

	// Try to determine format from the image content
	return readImageFormat(input)
}

// replaceExt replaces the file extension of inputName with ext
func replaceExt(inputName string, ext string) string {
	oldExt := filepath.Ext(inputName)
	return strings.TrimSuffix(inputName, oldExt) + "." + strings.TrimPrefix(ext, ".")
}

// processDirectoryImages handles the case when a directory of images is provided
func processDirectoryImages(flags config.GlobalSubCommandFlags) []ImageIO {
	inputFiles, err := GetImagesFromDirectoryRecursively(flags.InputDir)
	utils.HandleError(err)

	outputDir := config.GowallConfig.OutputFolder
	var operations []ImageIO

	for _, inputFile := range inputFiles {
		baseName := filepath.Base(inputFile.Path)
		ext, err := determineFileExt(flags, inputFile, nil)
		if err != nil {
			continue
		}
		outputPath := filepath.Join(outputDir, replaceExt(baseName, ext))
		operations = append(operations, ImageIO{
			ImageInput:  inputFile,
			ImageOutput: FileWriter{Path: outputPath},
			Format:      ext,
			Theme:       flags.Theme,
		})
	}

	return operations
}

// processBatchFiles handles the case when a list of input files is provided
func processBatchFiles(flags config.GlobalSubCommandFlags) []ImageIO {
	outputDir := config.GowallConfig.OutputFolder
	var operations []ImageIO

	for _, path := range flags.InputFiles {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		input := FileReader{Path: absolutePath}
		baseName := filepath.Base(absolutePath)
		ext, err := determineFileExt(flags, input, nil)
		if err != nil {
			continue
		}

		outputPath := filepath.Join(outputDir, replaceExt(baseName, ext))
		operations = append(operations, ImageIO{
			ImageInput:  input,
			ImageOutput: FileWriter{Path: outputPath},
			Format:      ext,
			Theme:       flags.Theme,
		})
	}

	return operations
}

// // // Then in your main processing function:
// func ProcessImages(operations []ImageIO) {
// 	for _, op := range operations {
// 		// Open input file
// 		inputData, err := loadImageData(op.ImageInput)
// 		if err != nil {
// 			utils.HandleError(err, "Failed to load input image")
// 			continue
// 		}
//
// 		// Process the image
// 		outputData, err := processImageWithTheme(inputData, op.Theme, op.Format)
// 		if err != nil {
// 			utils.HandleError(err, "Failed to process image")
// 			continue
// 		}
//
// 		// Save output (handle stdout specially)
// 		if op.ImageOutput == "/dev/stdout" {
// 			os.Stdout.Write(outputData)
// 		} else {
// 			err = saveImageData(outputData, op.ImageOutput)
// 			if err != nil {
// 				utils.HandleError(err, "Failed to save output image")
// 			}
// 		}
// 	}
// }
