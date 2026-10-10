package api

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Achno/gowall/config"
	request "github.com/Achno/gowall/pkg/requests"
	"github.com/PuerkitoBio/goquery"
)

type redditFeed struct {
	Entries []struct {
		Content string `xml:"content"`
	} `xml:"entry"`
}

func GetWallpaperOfTheDay() (string, error) {
	body, err := request.Get[[]byte](request.NewURLClient(0), config.WallOfTheDayUrl+"?t=day",
		request.WithHeader(request.Header{"User-Agent": "Mozilla/5.0 (compatible; gowall/1.0)"}),
	)
	if err != nil {
		var statusErr *request.StatusError
		if errors.As(err, &statusErr) {
			return "", fmt.Errorf("request failed with status code: %d %s", statusErr.Code, http.StatusText(statusErr.Code))
		}
		return "", err
	}

	var feed redditFeed
	if err := xml.Unmarshal(*body, &feed); err != nil {
		return "", fmt.Errorf("while reading the reddit feed: %w", err)
	}
	if len(feed.Entries) == 0 {
		return "", fmt.Errorf("could not find a post in the reddit feed")
	}

	post, err := goquery.NewDocumentFromReader(strings.NewReader(feed.Entries[0].Content))
	if err != nil {
		return "", err
	}
	if imageURL := firstPostImageURL(post.Selection); imageURL != "" {
		return imageURL, nil
	}
	return "", fmt.Errorf("there wasn't a top wallpaper today :( check later")
}

func firstPostImageURL(post *goquery.Selection) string {
	if src, ok := post.Find("img").Attr("src"); ok {
		if clean := cleanImageURL(src); isRedditImageURL(clean) {
			return clean
		}
	}

	// posts without a thumbnail still have a [link] to the image
	var found string
	post.Find("a[href]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if clean := cleanImageURL(s.AttrOr("href", "")); isRedditImageURL(clean) {
			found = clean
			return false
		}
		return true
	})
	return found
}

// cleanImageURL turns a preview.redd.it thumbnail into the full size i.redd.it image
func cleanImageURL(imageURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(imageURL))
	if err != nil {
		return ""
	}
	if strings.EqualFold(parsed.Hostname(), "preview.redd.it") {
		parsed.Host, parsed.RawQuery = "i.redd.it", ""
	}
	return parsed.String()
}

func isRedditImageURL(imageURL string) bool {
	parsed, err := url.Parse(imageURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "i.redd.it" || host == "external-preview.redd.it"
}
