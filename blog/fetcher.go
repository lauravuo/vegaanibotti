package blog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/lainio/err2/try"
	"golang.org/x/net/html"
)

type Post struct {
	ID          int64
	Title       string
	Description string
	URL         string
	Hashtags    []string
	Added       bool `json:"-"`
}

const classStr = "class"

func getTitleAndURL(tokenizer *html.Tokenizer, attrKey, attrValue string) (title, url string) {
	if attrKey == classStr && attrValue == "entry-title" {
		_ = tokenizer.Next() // a-tag
		_, moreAttr := tokenizer.TagName()

		var attrKeyBytes, attrValueBytes []byte

		for moreAttr {
			attrKeyBytes, attrValueBytes, moreAttr = tokenizer.TagAttr()
			if string(attrKeyBytes) == "href" {
				url = string(attrValueBytes)
				_ = tokenizer.Next() // a value

				return tokenizer.Token().Data, url
			}
		}
	}

	return "", ""
}

func getDescription(z *html.Tokenizer, attrKey, attrValue string) string {
	if attrKey == classStr && attrValue == "entry-summary" {
		_ = z.Next() // p-tag
		_ = z.Next() // a value

		return z.Token().Data
	}

	return ""
}

func getID(attrKey, attrValue string) int64 {
	if attrKey == classStr && strings.HasPrefix(attrValue, "teaser post-") {
		parts := strings.Split(attrValue, " ")
		strID, _ := strings.CutPrefix(parts[1], "post-")

		return try.To1(strconv.ParseInt(strID, 10, 64))
	}

	return 0
}

func loadExistingPosts(recipesFilePath string) (posts []Post, maxID int64) {
	// create file if it does not exist
	if _, err := os.Stat(recipesFilePath); errors.Is(err, os.ErrNotExist) {
		try.To(os.WriteFile(recipesFilePath, []byte("[]"), WritePerm))
	}

	// read existing posts
	fileContents := try.To1(os.ReadFile(recipesFilePath))
	try.To(json.Unmarshal(fileContents, &posts))

	if len(posts) > 0 {
		maxID = posts[0].ID
	}

	slog.Info("Existing posts", "count", len(posts), "maximum ID", maxID)

	return posts, maxID
}

func getAttributes(
	tokenizer *html.Tokenizer,
	moreAttr bool,
	maxID int64,
) (id int64, itemTitle, itemURL, description string, existingFound bool) {
	var attrKey, attrValue []byte

	var itemID int64

	for moreAttr {
		attrKey, attrValue, moreAttr = tokenizer.TagAttr()

		attrKeyStr := string(attrKey)
		attrValueStr := string(attrValue)

		if id := getID(attrKeyStr, attrValueStr); id != 0 {
			slog.Info("Handling post", "id", id)
			itemID = id

			if itemID <= maxID {
				existingFound = true

				break
			}
		}

		if title, url := getTitleAndURL(tokenizer, attrKeyStr, attrValueStr); title != "" {
			itemTitle = title
			itemURL = url
		}

		if desc := getDescription(tokenizer, attrKeyStr, attrValueStr); desc != "" {
			description = desc

			break
		}
	}

	return itemID, itemTitle, itemURL, description, existingFound
}

func FetchNewPosts(
	recipesFilePath string,
	httpGetter func(string, string) ([]byte, error),
) ([]Post, error) {
	posts, maxID := loadExistingPosts(recipesFilePath)
	existingFound := false

	url := "https://chocochili.net/luokka/paaruoat/page/%d/"
	index := 1

	added := make(map[int64]bool)

	for !existingFound {
		fetchURL := fmt.Sprintf(url, index)

		slog.Info("Fetching URL", "url", fetchURL)

		data, err := httpGetter(fetchURL, "")
		if err != nil {
			slog.Info("Stopped fetching", "count", index-1)

			break
		}

		tokenizer := html.NewTokenizer(bytes.NewReader(data))
		for !existingFound {
			tt := tokenizer.Next()

			if tt == html.ErrorToken {
				break
			}

			_, moreAttr := tokenizer.TagName()

			itemID, itemTitle, itemURL, description, existingFound := getAttributes(tokenizer, moreAttr, maxID)
			if !existingFound && description != "" {
				if _, ok := added[itemID]; !ok {
					posts = append(posts, Post{
						ID:          itemID,
						Title:       itemTitle,
						Description: description,
						URL:         itemURL,
						Added:       true,
					})

					added[itemID] = true
				}
			}
		}

		index++
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID > posts[j].ID
	})
	try.To(os.WriteFile(recipesFilePath, try.To1(json.Marshal(posts)), WritePerm))

	return posts, nil
}
