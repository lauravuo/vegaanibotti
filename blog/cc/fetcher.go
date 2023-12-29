package cc

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
	"github.com/lauravuo/vegaanibotti/blog/base"
	"golang.org/x/net/html"
)

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

func getID(attrKey, attrValue string) (id int64, tags []string) {
	if attrKey == classStr && strings.HasPrefix(attrValue, "teaser post-") {
		parts := strings.Split(attrValue, " ")
		strID, _ := strings.CutPrefix(parts[1], "post-")

		tags := make([]string, 0)

		for _, className := range parts {
			if strings.HasPrefix(className, "tag-") {
				tag, _ := strings.CutPrefix(className, "tag-")
				tags = append(tags, tag)
			}
		}

		return try.To1(strconv.ParseInt(strID, 10, 64)), tags
	}

	return 0, []string{}
}

func loadExistingPosts(recipesFilePath string) (posts []base.Post, maxID int64) {
	// create file if it does not exist
	if _, err := os.Stat(recipesFilePath); errors.Is(err, os.ErrNotExist) {
		try.To(os.WriteFile(recipesFilePath, []byte("[]"), base.WritePerm))
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

func getPost(tokenizer *html.Tokenizer, post *base.Post) {
	_, moreAttr := tokenizer.TagName()

	var attrKey, attrValue []byte

	for moreAttr {
		attrKey, attrValue, moreAttr = tokenizer.TagAttr()

		attrKeyStr := string(attrKey)
		attrValueStr := string(attrValue)

		if id, tagNames := getID(attrKeyStr, attrValueStr); id != 0 {
			post.ID = id
			post.Hashtags = tagNames
		}

		if title, url := getTitleAndURL(tokenizer, attrKeyStr, attrValueStr); title != "" {
			post.Title = title
			post.URL = url
		}

		if desc := getDescription(tokenizer, attrKeyStr, attrValueStr); desc != "" {
			post.Description = desc
		}
	}
}

func FetchNewPosts(
	recipesFilePath string,
	httpGetter func(string, string) ([]byte, error),
	previewOnly bool,
) (base.RecipeBank, error) {
	posts, maxID := loadExistingPosts(recipesFilePath)
	existingFound := false

	url := "https://chocochili.net/luokka/paaruoat/page/%d/"
	index := 1

	added := make(map[int64]bool)

	post := &base.Post{}

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

			getPost(tokenizer, post)

			if post.IsValid() {
				existingFound = post.ID <= maxID
				if !existingFound {
					if _, ok := added[post.ID]; !ok {
						tags := post.Hashtags
						post.Hashtags = make([]string, 0)
						post.Hashtags = append(post.Hashtags, []string{"chocochili", "vegaani", "vegaaniresepti"}...)
						post.Hashtags = append(post.Hashtags, tags...)
						post.Added = true
						posts = append(posts, *post)

						slog.Info("Added new post",
							"ID", post.ID,
							"Title", post.Title,
							"Description", post.Description,
							"URL", post.URL,
							"Hashtags", post.Hashtags,
						)

						added[post.ID] = true
						post = &base.Post{}
					}
				}
			}
		}

		index++
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID > posts[j].ID
	})

	if !previewOnly {
		try.To(os.WriteFile(recipesFilePath, try.To1(json.Marshal(posts)), base.WritePerm))
	}

	return base.RecipeBank{
		Posts:       posts,
		UsedIDsPath: UsedIDsPath,
	}, nil
}
