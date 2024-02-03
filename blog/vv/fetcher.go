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

//nolint:gocyclo,gocognit,cyclop
func getPost(tokenizer *html.Tokenizer) *base.Post {
	tagName, moreAttr := tokenizer.TagName()

	const articleStr = "article"

	var attrKey, attrValue []byte

	isVegan := false
	isTip := false

	var postID int64

	for moreAttr {
		attrKey, attrValue, moreAttr = tokenizer.TagAttr()

		attrKeyStr := string(attrKey)
		attrValueStr := string(attrValue)

		if string(tagName) == articleStr && attrKeyStr == classStr {
			parts := strings.Split(attrValueStr, " ")

			for _, part := range parts {
				if part == "tag-vegaani" {
					isVegan = true
				} else if part == "tag-koosteet" || part == "category-vinkit" {
					isTip = true
				} else if strings.HasPrefix(part, "post-") {
					postID, _ = strconv.ParseInt(strings.ReplaceAll(part, "post-", ""), 10, 64)
				}
			}
		}
	}

	//nolint:nestif
	if isVegan && !isTip {
		title := ""
		postURL := ""

		for title == "" && postURL == "" {
			tt := tokenizer.Next()
			tagName, _ := tokenizer.TagName()

			if tt == html.ErrorToken || (tt == html.EndTagToken && string(tagName) == articleStr) {
				break
			}

			attrKey, attrValue, _ = tokenizer.TagAttr()
			if string(tagName) == "h2" && string(attrKey) == classStr && strings.HasPrefix(string(attrValue), "entry-title") {
				tagName, _ := tokenizer.TagName()
				for string(tagName) != "a" {
					tt := tokenizer.Next()
					tagName, _ = tokenizer.TagName()

					if tt == html.ErrorToken || (tt == html.EndTagToken && string(tagName) == articleStr) {
						break
					}
				}

				if string(tagName) == "a" {
					moreAttr := true
					for moreAttr {
						attrKey, attrValue, moreAttr = tokenizer.TagAttr()
						if string(attrKey) == "href" {
							postURL = string(attrValue)
							_ = tokenizer.Next() // a value

							title = tokenizer.Token().Data

							break
						}
					}
				}
			}
		}

		return &base.Post{
			ID:    postID,
			Title: title,
			URL:   postURL,
		}
	}

	return nil
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

			post := getPost(tokenizer)

			//nolint:nestif
			if post != nil && post.IsValid() {
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
