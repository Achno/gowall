package api

import (
	"fmt"
	"net/http"

	"github.com/Achno/gowall/config"
	"github.com/PuerkitoBio/goquery"
)

func GetWallpaperOfTheDay() (string, error) {
	url := config.WallOfTheDayUrl

	// http.Get() doesn't send a User-Agent header, Reddit blocks it with a 429
	// so we build the request manually so we can add one
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Reddit blocks requests that don't look like they're coming from a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; gowall/1.0)")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status code: %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	defer response.Body.Close()

	// Parse the html and select the top wallpaper of the day
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", err
	}

	var imageUrls []string

	doc.Find("img.i18n-post-media-img").Each(func(index int, selection *goquery.Selection) {
		imgUrl, exists := selection.Attr("src")

		if exists {
			imageUrls = append(imageUrls, imgUrl)
		}
	})

	// if no posts were found
	if len(imageUrls) == 0 {
		return "", fmt.Errorf("there wasn't a top wallpaper today :( check later")
	}

	// we return the second url : example https://i.redd.it/xhupkp01d5id1.png with correct dimensions instead of the preview img
	return imageUrls[1], nil
}
