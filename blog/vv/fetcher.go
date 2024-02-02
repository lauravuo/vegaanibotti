package vv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	"golang.org/x/net/html"
)

const RecipesPath = base.DataPath + "/vv/recipes.json"

const UsedIDsPath = base.DataPath + "/vv/used.json"

const classStr = "class"

func getTitleAndURL(tokenizer *html.Tokenizer, attrKey, attrValue string) (title, postURL string) {
	if attrKey == classStr && strings.HasPrefix(attrValue, "entry-title") {
		var tagName []byte

		var moreAttr bool

		for len(tagName) == 0 {
			tt := tokenizer.Next() // a-tag
			if tt == html.ErrorToken {
				break
			}

			tagName, moreAttr = tokenizer.TagName()
		}

		var attrKeyBytes, attrValueBytes []byte

		for moreAttr {
			attrKeyBytes, attrValueBytes, moreAttr = tokenizer.TagAttr()
			if string(attrKeyBytes) == "href" {
				postURL = string(attrValueBytes)
				_ = tokenizer.Next() // a value

				return tokenizer.Token().Data, postURL
			}
		}
	}

	return "", ""
}

func getDescription(_ *html.Tokenizer, _, _ string) string {
	return ""
}

func getImages(_ *html.Tokenizer, _, _ string) (thumbnail, image string) {
	return "", ""
}

func getID(tagName, attrKey, attrValue string) (postID int64, isVegan, isTip bool) {
	if tagName == "article" && attrKey == classStr {
		parts := strings.Split(attrValue, " ")

		for _, part := range parts {
			if part == "tag-vegaani" {
				isVegan = true
			} else if part == "tag-koosteet" || part == "category-vinkit" {
				isTip = true
			} else if strings.HasPrefix(part, "post-") {
				postID, _ = strconv.ParseInt(strings.ReplaceAll(part, "post-", ""), 10, 64)
			}
		}

		if postID > 0 {
			return postID, isVegan, isTip
		}
	}

	return 0, isVegan, isTip
}

func getPost(tokenizer *html.Tokenizer, post *base.Post) {
	tagName, moreAttr := tokenizer.TagName()

	var attrKey, attrValue []byte

	for moreAttr {
		attrKey, attrValue, moreAttr = tokenizer.TagAttr()

		attrKeyStr := string(attrKey)
		attrValueStr := string(attrValue)

		if id, isVegan, isTip := getID(string(tagName), attrKeyStr, attrValueStr); id != 0 {
			post.ID = id
			post.Hashtags = []string{}

			if !isVegan || isTip {
				post.Hashtags = append(post.Hashtags, "remove")
			}
		}

		if title, postURL := getTitleAndURL(tokenizer, attrKeyStr, attrValueStr); title != "" {
			post.Title = title
			post.URL = postURL
		}

		if desc := getDescription(tokenizer, attrKeyStr, attrValueStr); desc != "" {
			post.Description = desc
		}

		if thumbnail, image := getImages(tokenizer, attrKeyStr, attrValueStr); thumbnail != "" {
			post.ThumbnailURL = thumbnail
			post.ImageURL = image
		}
	}
}

//nolint:cyclop,gocognit
func fetchPosts(
	httpGetter func(string, string) ([]byte, error),
	maxID int64,
) []base.Post {
	posts := make([]base.Post, 0)
	existingFound := false

	// fetch all after maxID
	baseURL := "https://vegeviettelys.fi/tag/kasvisruoat/page/%d/"

	index := 1

	added := make(map[int64]bool)

	post := &base.Post{}

	for !existingFound {
		fetchURL := fmt.Sprintf(baseURL, index)
		slog.Info("Fetching URL", "url", fetchURL)

		data, err := httpGetter(
			fetchURL,
			"",
		)
		if err != nil || len(data) == 0 {
			slog.Info("Stopped fetching", "round", index-1)

			break
		}

		tokenizer := html.NewTokenizer(bytes.NewReader(data))
		for !existingFound {
			tt := tokenizer.Next()

			if tt == html.ErrorToken {
				break
			}

			getPost(tokenizer, post)

			//nolint:nestif
			if post.IsValid() {
				existingFound = post.ID <= maxID
				if !existingFound {
					if _, ok := added[post.ID]; !ok {
						tags := post.Hashtags
						post.Hashtags = make([]string, 0)
						post.Hashtags = append(post.Hashtags, []string{"vegeviettelys", "vegaani", "vegaaniresepti"}...)
						post.Hashtags = append(post.Hashtags, tags...)
						post.Added = true

						if len(tags) == 0 || tags[0] != "remove" {
							posts = append(posts, *post)
							slog.Info("Added new post",
								"ID", post.ID,
								"Title", post.Title,
								"Description", post.Description,
								"URL", post.URL,
								"Thumbnail", post.ThumbnailURL,
								"Image", post.ImageURL,
								"Hashtags", post.Hashtags,
							)

							added[post.ID] = true
						}

						post = &base.Post{}
					}
				}

				if existingFound {
					break
				}
			}
		}

		index++
	}

	return posts
}

func FetchNewPosts(
	recipesFilePath string,
	httpGetter func(string, string) ([]byte, error),
	_ func(string, url.Values, string) (data []byte, err error),
	previewOnly bool,
) (base.RecipeBank, error) {
	posts, maxID := base.LoadExistingPosts(recipesFilePath)

	mainPosts := fetchPosts(httpGetter, maxID)
	if len(mainPosts) > 0 {
		sort.Slice(mainPosts, func(i, j int) bool {
			return mainPosts[i].ID > mainPosts[j].ID
		})

		posts = append(mainPosts, posts...)
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
