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
	Added       bool `json:"-"`
}

const classStr = "class"

func getTitleAndURL(z *html.Tokenizer, attrKey, attrValue string) (string, string) {
	if attrKey == classStr && attrValue == "entry-title" {
		_ = z.Next() // a-tag
		_, moreAttr := z.TagName()

		var attrKeyBytes, attrValueBytes []byte

		for moreAttr {
			attrKeyBytes, attrValueBytes, moreAttr = z.TagAttr()
			if string(attrKeyBytes) == "href" {
				url := string(attrValueBytes)
				_ = z.Next() // a value

				return z.Token().Data, url
			}
		}
	}

	return "", ""
}

func getDescription(z *html.Tokenizer, attrKey, attrValue string) string {
	if attrKey == classStr && attrValue == "entry-summary" {
		_ = z.Next() // p-tag
		_ = z.Next() // a value
		return string(z.Token().Data)
	}
	return ""
}

func getID(z *html.Tokenizer, attrKey, attrValue string) int64 {
	if attrKey == classStr && strings.HasPrefix(attrValue, "teaser post-") {
		parts := strings.Split(attrValue, " ")
		strID, _ := strings.CutPrefix(parts[1], "post-")
		return try.To1(strconv.ParseInt(strID, 10, 64))
	}
	return 0
}

func loadExistingPosts(recipesFilePath string) ([]Post, int64) {
	// create file if it does not exist
	if _, err := os.Stat(recipesFilePath); errors.Is(err, os.ErrNotExist) {
		try.To(os.WriteFile(recipesFilePath, []byte("[]"), 0644))
	}

	// read existing posts
	posts := make([]Post, 0)
	fileContents := try.To1(os.ReadFile(recipesFilePath))
	try.To(json.Unmarshal(fileContents, &posts))
	var maxID int64
	if len(posts) > 0 {
		maxID = posts[0].ID
	}
	slog.Info("Existing posts", "count", len(posts), "maximum ID", maxID)
	return posts, maxID
}

func FetchNewPosts(
	recipesFilePath string,
	httpGetter func(string, string) ([]byte, error),
) ([]Post, error) {

	posts, maxID := loadExistingPosts(recipesFilePath)
	existingFound := false

	url := "https://chocochili.net/luokka/paaruoat/page/%d/"
	index := 1
	var itemID int64
	var itemTitle, itemURL string
	var attrKey, attrValue []byte
	added := make(map[int64]bool)
	for !existingFound {
		fetchUrl := fmt.Sprintf(url, index)
		slog.Info("Fetching URL", "url", fetchUrl)

		data, err := httpGetter(fetchUrl, "")
		if err != nil {
			slog.Info("Stopped fetching", "count", index-1)
			break
		}
		z := html.NewTokenizer(bytes.NewReader(data))
		for !existingFound {
			tt := z.Next()
			if tt == html.ErrorToken {
				break
			}

			_, moreAttr := z.TagName()
			for moreAttr {
				attrKey, attrValue, moreAttr = z.TagAttr()
				attrKeyStr := string(attrKey)
				attrValueStr := string(attrValue)
				if id := getID(z, attrKeyStr, attrValueStr); id != 0 {
					slog.Info("Handling post", "id", id)
					itemID = id
					if itemID <= maxID {
						existingFound = true
						break
					}
				}
				if title, url := getTitleAndURL(z, attrKeyStr, attrValueStr); title != "" {
					itemTitle = title
					itemURL = url
				}
				if desc := getDescription(z, attrKeyStr, attrValueStr); desc != "" {
					if _, ok := added[itemID]; !ok {
						posts = append(posts, Post{
							ID:          itemID,
							Title:       itemTitle,
							Description: desc,
							URL:         itemURL,
							Added:       true,
						})
						added[itemID] = true
					}
				}
			}
		}
		index += 1
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID > posts[j].ID
	})
	try.To(os.WriteFile(recipesFilePath, try.To1(json.Marshal(posts)), 0644))
	return posts, nil
}
