package api

import (
	"fmt"
	"net/http"

	"github.com/Achno/gowall/config"
	"github.com/PuerkitoBio/goquery"
)

func GetWallpaperOfTheDay() (string, error) {
	url := config.WallOfTheDayUrl

	if url == "" {
		return "", fmt.Errorf("WallOfTheDayUrl is not configured")
	}

	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch the URL: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status code: %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML document: %w", err)
	}

	var imageUrls []string
	doc.Find("img.i18n-post-media-img").Each(func(index int, selection *goquery.Selection) {
		if imgUrl, exists := selection.Attr("src"); exists {
			imageUrls = append(imageUrls, imgUrl)
		}
	})

	// Check if any images were found
	if len(imageUrls) < 2 {
		return "", fmt.Errorf("no suitable wallpapers found today, try again later")
	}

	// Return the second URL (assuming it is of better quality)
	return imageUrls[1], nil
}
