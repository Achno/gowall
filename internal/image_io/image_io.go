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
	"github.com/spf13/cobra"
)

type ImageIO struct {
	ImageInput  ImageReader
	ImageOutput ImageWriter
	Format      string
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
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fr FileReader) String() string {
	filePath, _ := filepath.Abs(fr.Path)
	return filePath
}

func (fw FileWriter) Create() (*os.File, error) {
	dir := filepath.Dir(fw.Path)

	// Create all necessary parent directories
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

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
func DetermineImageOperations(flags config.GlobalSubCommandFlags, args []string, cmd *cobra.Command) ([]ImageIO, error) {
	// Process by priority: directory > batch files > single file/stdin
	if flags.InputDir != "" {
		imgIO, err := directoryIO(flags, cmd)
		if err != nil {
			return nil, err
		}
		return imgIO, nil
	}

	if len(flags.InputFiles) > 0 {
		return batchIO(flags, cmd), nil
	}

	imgIO, err := SingleIO(flags, args, cmd)
	if err != nil {
		return nil, err
	}
	return imgIO, nil
}

// SingleIO handles both file and STDIN input cases
func SingleIO(flags config.GlobalSubCommandFlags, args []string, cmd *cobra.Command) ([]ImageIO, error) {
	input := determineInput(args)
	output, ext, err := determineOutput(flags, args, input, cmd)
	if err != nil {
		return nil, fmt.Errorf("could not determine output: ")
	}
	return []ImageIO{
		{
			ImageInput:  input,
			ImageOutput: output,
			Format:      ext,
		},
	}, nil
}

// directoryIO handles the case when a directory of images is provided
func directoryIO(flags config.GlobalSubCommandFlags, cmd *cobra.Command) ([]ImageIO, error) {

	filter := func(path string, entry fs.DirEntry) bool {
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if cmd.Name() == "ocr" {
			return config.SupportedTextExtensions[ext] || config.SupportedImageExtensions[ext]
		}
		return config.SupportedImageExtensions[ext]
	}

	inputFiles, err := GetFilesFromDirectory(flags.InputDir, filter)
	if err != nil {
		return nil, err
	}
	var operations []ImageIO
	dir := config.GowallConfig.OutputFolder

	// --output - multiple files
	if flags.OutputDestination != "" {
		dir = filepath.Join(flags.OutputDestination)
	}

	for _, inputFile := range inputFiles {
		baseName := filepath.Base(inputFile.Path)
		ext, err := determineFileExt(flags, inputFile, nil, cmd)
		if err != nil {
			continue
		}
		outputPath := filepath.Join(dir, replaceExt(baseName, ext))
		operations = append(operations, ImageIO{
			ImageInput:  inputFile,
			ImageOutput: FileWriter{Path: outputPath},
			Format:      ext,
		})
	}

	return operations, nil
}

// batchIO handles the case when a list of input files is provided
func batchIO(flags config.GlobalSubCommandFlags, cmd *cobra.Command) []ImageIO {
	var operations []ImageIO
	dir := config.GowallConfig.OutputFolder

	// --output - multiple files
	if flags.OutputDestination != "" {
		dir = filepath.Join(flags.OutputDestination)
	}

	// expand the tilde (~) to the full path in case the shell does not
	files := utils.ExpandTilde(flags.InputFiles)

	for _, path := range files {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		input := FileReader{Path: absolutePath}
		baseName := filepath.Base(absolutePath)
		ext, err := determineFileExt(flags, input, nil, cmd)
		if err != nil {
			continue
		}

		outputPath := filepath.Join(dir, replaceExt(baseName, ext))
		operations = append(operations, ImageIO{
			ImageInput:  input,
			ImageOutput: FileWriter{Path: outputPath},
			Format:      ext,
		})
	}

	return operations
}

// determineInput resolves the input source (file or stdin)
func determineInput(args []string) ImageReader {
	// If the first arg is "-", use stdin
	if len(args) > 0 && args[0] == "-" {
		return Stdin{}
	}

	// Otherwise file
	f := utils.ExpandTilde(args)
	return FileReader{Path: f[0]}
}

// determineOutput resolves the output destination and format
func determineOutput(flags config.GlobalSubCommandFlags, args []string, input ImageReader, cmd *cobra.Command) (ImageWriter, string, error) {
	// Check if output should be stdout
	if IsStdoutOutput(flags, args) {
		ext, err := determineFileExt(flags, input, Stdout{}, cmd)
		if err != nil {
			return nil, "", err
		}
		return Stdout{}, ext, nil
	}

	outputDest, err := resolveOutputPath(flags, args, input, cmd)
	if err != nil {
		return nil, "", err
	}
	output := FileWriter{Path: outputDest}
	ext, err := determineFileExt(flags, input, output, cmd)
	if err != nil {
		return nil, "", err
	}

	return output, ext, nil
}

// resolveOutputPath determines the final output path based on flags and args
func resolveOutputPath(flags config.GlobalSubCommandFlags, args []string, input ImageReader, cmd *cobra.Command) (string, error) {
	dir := config.GowallConfig.OutputFolder
	name, err := generateFileName(flags, args, input, cmd)
	if err != nil {
		return "", err
	}

	// --output full destination - single file
	if flags.OutputDestination != "" && filepath.Ext(flags.OutputDestination) != "" && (flags.InputDir == "" && len(flags.InputFiles) <= 0) {
		return flags.OutputDestination, nil
	}

	// --output directory - single file
	if flags.OutputDestination != "" && filepath.Ext(flags.OutputDestination) == "" && (flags.InputDir == "" && len(flags.InputFiles) <= 0) {
		dir = flags.OutputDestination
		return filepath.Join(dir, name), nil
	}

	return filepath.Join(dir, name), nil
}

// generateFileName creates a filename with an extension for an image
func generateFileName(flags config.GlobalSubCommandFlags, args []string, input ImageReader, cmd *cobra.Command) (string, error) {
	// For stdin input, generate timestamp-based filename
	if len(args) > 0 && args[0] == "-" {
		ts := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("img-%s", ts)
		ext, err := determineFileExt(flags, input, nil, cmd)
		if err != nil {
			return "", err
		}
		return filepath.Join(filename + "." + ext), nil
	}

	// For file input, base output on input filename
	absInput, err := filepath.Abs(args[0])
	utils.HandleError(err, "could not resolve absolute path for input")
	baseName := filepath.Base(absInput)
	ext, err := determineFileExt(flags, input, nil, cmd)
	if err != nil {
		return "", err
	}
	return filepath.Join(replaceExt(baseName, ext)), nil
}

// IsStdoutOutput checks if the output destination indicates stdout
func IsStdoutOutput(flags config.GlobalSubCommandFlags, args []string) bool {
	return flags.OutputDestination == "-" ||
		flags.OutputDestination == "/dev/stdout" ||
		(len(args) > 1 && args[1] == "-")
}

// Determines file extension based on flags and the arguments, will return "png" if nothing is satisfied
// make the cobra arguement optional, varidadic
func determineFileExt(flags config.GlobalSubCommandFlags, input ImageReader, output ImageWriter, cmd *cobra.Command) (string, error) {
	// Ext from --format flag
	if flags.Format != "" {
		return flags.Format, nil
	}

	// Ext from --output flag
	if ext := filepath.Ext(flags.OutputDestination); ext != "" {
		return strings.ReplaceAll(ext, ".", ""), nil
	}

	// if 'gowall ocr' is invoked make the default 'md' and then listen for the --format flag
	if cmd != nil && cmd.Name() == "ocr" {
		if flags.Format != "" {
			return flags.Format, nil
		}
		return "md", nil
	}

	// Check if output is a FileWriter to get its path
	if fileWriter, ok := output.(FileWriter); ok {
		if ext := filepath.Ext(fileWriter.Path); ext != "" {
			return strings.ReplaceAll(ext, ".", ""), nil
		}
	}

	// Ext from a Readers Source
	if fileReader, ok := input.(FileReader); ok {
		if ext := filepath.Ext(fileReader.Path); ext != "" {
			return strings.ReplaceAll(ext, ".", ""), nil
		}
	}

	//? If there is a file in stdin assume its a png, so it gets encoded later
	if _, ok := input.(Stdin); ok {
		return "png", nil
	}

	return "", fmt.Errorf("extension not found")
}

// replaceExt replaces the file extension of inputName with ext
func replaceExt(inputName string, ext string) string {
	oldExt := filepath.Ext(inputName)
	return strings.TrimSuffix(inputName, oldExt) + "." + strings.TrimPrefix(ext, ".")
}

func GetFilesFromDirectory(path string, filter func(string, fs.DirEntry) bool) ([]FileReader, error) {
	var files []FileReader
	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filter(path, entry) {
			files = append(files, FileReader{Path: path})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in directory or subdirectories")
	}

	return files, nil
}

func LoadFileBytes(src ImageReader) ([]byte, error) {
	reader, err := src.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	bytes, err := io.ReadAll(reader)

	return bytes, err
}
