package image

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Achno/gowall/utils"

	"github.com/chai2010/webp"
)

// Available formats to Encode an image in
var encoders = map[string]func(file *os.File, img image.Image) error{
    "png": func(file *os.File, img image.Image) error {
        return png.Encode(file, img)
    },
    "jpg": func(file *os.File, img image.Image) error {
        return jpeg.Encode(file, img, nil)
    },
    "jpeg": func(file *os.File, img image.Image) error {
        return jpeg.Encode(file, img, nil)
    },
    "webp": func(file *os.File, img image.Image) error {
        return webp.Encode(file, img, nil)
    },
}

func LoadImage(filePath string) (image.Image , error){
	
	file,err := os.Open(filePath)

	if err != nil {
		return nil,err
	}

	defer file.Close()

	img,_,err := image.Decode(file)

	return img,err
}

func SaveImage(img image.Image, filePath string, format string) error{

	file,err := os.Create(filePath)

	if err != nil {
		return err
	}

	defer file.Close()

	
    encoder, ok := encoders[strings.ToLower(format)]

    if !ok {
        return fmt.Errorf("unsupported format: %s", format)
    }

    return encoder(file, img)

}

func ProcessImg(imgPath string, theme string ) error {

	img, err := LoadImage(imgPath)

	if err != nil {
		fmt.Println("Error loading image :", err)
		return err
	}

	var selectedTheme Theme

	selectedTheme , err = SelectTheme(theme)

	if err != nil{
		fmt.Println("Unknown theme:", theme)
		return err
	}

	newImg, err := convertImage(img,selectedTheme)

	if err != nil {
		fmt.Println("Error Converting image:", err)
		return err
	}

	//Extract file extension from imgPath
	extension := strings.ToLower(filepath.Ext(imgPath))

	if extension == "" {
		fmt.Println("Error: Could not determine file extension.")
		return err
	}

	// remove '.' from the extension
	extension = extension[1:]

	dirPath, err := utils.CreateDirectory()
	nameOfFile := filepath.Base(imgPath)
	
    outputFilePath := filepath.Join(dirPath, nameOfFile)


	if err != nil{
		fmt.Println("Error creating Directory or getting path")
		return err
	}

	err = SaveImage(newImg, outputFilePath, extension)

	if err != nil {
		fmt.Println("Error saving image:", err, outputFilePath)
		return err
	}

	fmt.Printf("Image processed and saved as %s\n", outputFilePath)
	
	return nil

}

func ProcessBatchImgs(files []string , theme string){

	var wg sync.WaitGroup

	for index, file := range files{
		
		wg.Add(1)

		go func (file string, index int)  {
			defer wg.Done()

			ok :=ProcessImg(file,theme)

			if ok != nil {
				os.Exit(1)
			}

			fmt.Printf(" ::: Image %d Completed , %d Images left ::: \n",index,len(files) -index -1)
		}(file,index)

	}

	wg.Wait()
}
