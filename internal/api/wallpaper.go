package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Achno/gowall/config"
)

type redditListing struct {
	Data struct {
		Children []struct {
			Data redditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type redditPost struct {
	PostHint            string `json:"post_hint"`
	URL                 string `json:"url"`
	URLOverriddenByDest string `json:"url_overridden_by_dest"`
	Over18              bool   `json:"over_18"`
	IsVideo             bool   `json:"is_video"`
}

func GetWallpaperOfTheDay() (string, error) {
	req, err := http.NewRequest("GET", config.WallOfTheDayUrl+".json?t=day&limit=10&raw_json=1", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; gowall/1.0)")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status code: %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	var listing redditListing
	if err := json.NewDecoder(response.Body).Decode(&listing); err != nil {
		return "", err
	}

	for _, child := range listing.Data.Children {
		post := child.Data

		if post.Over18 || post.IsVideo || post.PostHint != "image" {
			continue
		}

		if post.URLOverriddenByDest != "" {
			return post.URLOverriddenByDest, nil
		}

		if post.URL != "" {
			return post.URL, nil
		}
	}

	return "", fmt.Errorf("reddit returned no image posts in the top wallpaper feed")
}
