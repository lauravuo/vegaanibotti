package vmm

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

const RecipesPath = base.DataPath + "/vmm/recipes.json"

const UsedIDsPath = base.DataPath + "/vmm/used.json"

const classStr = "class"

func getTitleAndURL(tokenizer *html.Tokenizer, attrKey, attrValue string) (title, url string) {
	if attrKey == classStr && strings.HasPrefix(attrValue, "entry-title") {
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
	return ""
}

func getImages(tokenizer *html.Tokenizer, attrKey, attrValue string) (thumbnail, image string) {
	if attrKey == "data-bgset" {
		return attrValue, attrValue
	}
	return thumbnail, image
}

func getID(tagName, attrKey, attrValue string) (id int64) {
	if tagName == "article" && attrKey == "id" && strings.HasPrefix(attrValue, "post-") {
		strID := strings.ReplaceAll(attrValue, "post-", "")

		return try.To1(strconv.ParseInt(strID, 10, 64))
	}

	return 0
}

func getPost(tokenizer *html.Tokenizer, post *base.Post) {
	tagName, moreAttr := tokenizer.TagName()

	var attrKey, attrValue []byte

	for moreAttr {
		attrKey, attrValue, moreAttr = tokenizer.TagAttr()

		attrKeyStr := string(attrKey)
		attrValueStr := string(attrValue)

		if id := getID(string(tagName), attrKeyStr, attrValueStr); id != 0 {
			post.ID = id
			post.Hashtags = []string{}
		}

		if title, url := getTitleAndURL(tokenizer, attrKeyStr, attrValueStr); title != "" {
			post.Title = title
			post.URL = url
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

func fetchPostsForCategory(category string) []base.Post {
	// fetch all after maxID
}

func FetchNewPosts(
	recipesFilePath string,
	httpGetter func(string, string) ([]byte, error),
	httpPoster func(string, url.Values, string) (data []byte, err error),
	previewOnly bool,
) (base.RecipeBank, error) {
	posts, maxID := base.LoadExistingPosts(recipesFilePath)
	existingFound := false

	fetchURL := "https://viimeistamuruamyoten.com/wp-admin/admin-ajax.php"

	params := url.Values{}
	params.Add("order", "desc")
	params.Add("layout", "photography")
	params.Add("from", "customize")
	params.Add("template", "sidebar")
	params.Add("ppp", "6")
	params.Add("archivetype", "cat")
	params.Add("archivevalue", "177") // vegan 232
	params.Add("action", "penci_archive_more_post_ajax")
	params.Add("nonce", "50fe87c6df")

	index := 1

	added := make(map[int64]bool)

	post := &base.Post{}

	for !existingFound {
		slog.Info("Fetching URL", "url", fetchURL)

		params.Set("offset", fmt.Sprintf("%d", len(added)))

		data, err := httpPoster(
			fetchURL,
			params,
			"",
		)
		if err != nil {
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

			if post.IsValid() {
				existingFound = post.ID <= maxID
				if !existingFound {
					if _, ok := added[post.ID]; !ok {
						tags := post.Hashtags
						post.Hashtags = make([]string, 0)
						post.Hashtags = append(post.Hashtags, []string{"viimeistämuruamyöden", "vegaani", "vegaaniresepti"}...)
						post.Hashtags = append(post.Hashtags, tags...)
						post.Added = true
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
