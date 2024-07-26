package image

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
)

type ThemeConverter struct {

}

func (themeConv *ThemeConverter) Process(img image.Image, theme string) (image.Image, error){

	selectedTheme, err := SelectTheme(theme)

	if err != nil{
		fmt.Println("Unknown theme:", theme)
		return nil,err
	}

	newImg, err := convertImage(img,selectedTheme)

	if err != nil {
		fmt.Println("Error Converting image:", err)
		return nil,err
	}

	return newImg,nil
}

func convertImage(img image.Image, theme Theme ) (image.Image, error){
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)


	// replace each pixel with the selected theme's nearest color
    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
        for x := bounds.Min.X; x < bounds.Max.X; x++ {
            originalColor := img.At(x, y)
            newColor := nearestColor(originalColor, theme)
            newImg.Set(x, y, newColor)
        }
    }
	
	if newImg == nil {return nil, errors.New("error processing the Image")} 


	return newImg, nil

}

func nearestColor(clr color.Color, theme Theme) color.Color{

	r,g,b,_ := clr.RGBA()

	// Convert from 16-bit to 8-bit
	r, g, b = r>>8, g>>8, b>>8


	minDist := math.MaxFloat64

	var nearestClr color.Color

	for _ ,themeColor := range theme.Colors{
		tr , tg , tb , _ := themeColor.RGBA()
		// Convert from 16-bit to 8-bit
		tr, tg, tb = tr>>8, tg>>8, tb>>8

		distance := colorDistance(tr,tg,tb,r,g,b)

		if distance < minDist{
			minDist = distance
			nearestClr = themeColor
		}

	}

	return nearestClr

}



func colorDistance(r1, g1, b1, r2, g2, b2 uint32) float64 {
    return math.Sqrt(float64((r1-r2)*(r1-r2) + (g1-g2)*(g1-g2) + (b1-b2)*(b1-b2)))
}
