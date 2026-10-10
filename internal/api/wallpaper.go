package api

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/Achno/gowall/config"
	request "github.com/Achno/gowall/pkg/requests"
	"github.com/PuerkitoBio/goquery"
)

func GetWallpaperOfTheDay() (string, error) {
	body, err := request.Get[[]byte](request.NewURLClient(0), config.WallOfTheDayUrl+"?sort=top&t=day",
		request.WithHeader(request.Header{
			"User-Agent": "Mozilla/5.0 (compatible; gowall/1.0)",
			"Accept":     "text/html",
		}),
	)
	if err != nil {
		var statusErr *request.StatusError
		if errors.As(err, &statusErr) {
			return "", fmt.Errorf("request failed with status code: %d %s", statusErr.Code, http.StatusText(statusErr.Code))
		}
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(*body))
	if err != nil {
		return "", err
	}
	post := firstPostSelection(doc)
	if post == nil {
		return "", fmt.Errorf("could not find a post element on the page")
	}
	if imageURL := firstPostImageURL(post); imageURL != "" {
		return imageURL, nil
	}
	return "", fmt.Errorf("there wasn't a top wallpaper today :( check later")
}

func firstPostSelection(doc *goquery.Document) *goquery.Selection {
	for _, selector := range []string{"shreddit-post", "article", "[data-testid='post-container']", ".thing"} {
		if post := doc.Find(selector).First(); post.Length() > 0 {
			return post
		}
	}
	return nil
}

func firstPostImageURL(post *goquery.Selection) string {
	imageAttrs := []string{"src", "data-src"}

	if u, ok := post.Attr("data-url"); ok {
		if clean := cleanImageURL(u); isRedditImageURL(clean) {
			return clean
		}
	}

	var found string
	post.Find("img, source").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if srcset, ok := s.Attr("srcset"); ok {
			if u := bestFromSrcset(srcset); isRedditImageURL(u) {
				found = u
				return false
			}
		}
		for _, attr := range imageAttrs {
			if u, ok := s.Attr(attr); ok {
				if clean := cleanImageURL(u); isRedditImageURL(clean) {
					found = clean
					return false
				}
			}
		}
		return true
	})
	if found != "" {
		return found
	}

	post.Find("a[href]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if u, ok := s.Attr("href"); ok {
			if clean := cleanImageURL(u); isRedditImageURL(clean) {
				found = clean
				return false
			}
		}
		return true
	})
	return found
}

func bestFromSrcset(srcset string) string {
	best, bestW := "", 0
	for _, candidate := range strings.Split(srcset, ",") {
		fields := strings.Fields(strings.TrimSpace(candidate))
		if len(fields) == 0 {
			continue
		}
		w := 0
		if len(fields) > 1 {
			fmt.Sscanf(fields[1], "%dw", &w)
		}
		if clean := cleanImageURL(fields[0]); w > bestW && isRedditImageURL(clean) {
			best, bestW = clean, w
		}
	}
	return best
}

func cleanImageURL(imageURL string) string {
	imageURL = strings.TrimSpace(html.UnescapeString(imageURL))
	if imageURL == "" || strings.HasPrefix(imageURL, "data:") || strings.HasPrefix(imageURL, "blob:") {
		return ""
	}
	if strings.HasPrefix(imageURL, "//") {
		imageURL = "https:" + imageURL
	}
	parsed, err := url.Parse(imageURL)
	if err != nil {
		return imageURL
	}
	if strings.EqualFold(parsed.Hostname(), "preview.redd.it") {
		parsed.Scheme, parsed.Host, parsed.RawQuery, parsed.ForceQuery = "https", "i.redd.it", "", false
		return parsed.String()
	}
	return imageURL
}

func isRedditImageURL(imageURL string) bool {
	parsed, err := url.Parse(imageURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "i.redd.it" || host == "preview.redd.it" || host == "external-preview.redd.it"
}
