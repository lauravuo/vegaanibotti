package kk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
)

const RecipesPath = base.DataPath + "/kk/recipes.json"

const UsedIDsPath = base.DataPath + "/kk/used.json"

var errFeed = errors.New("unable to parse feed")

// slugToID derives a stable int64 ID from a slug using FNV-1a hash.
func slugToID(slug string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(slug))
	return int64(h.Sum64() & 0x7fffffffffffffff) // keep positive
}

type Image struct {
	Filename string `json:"filename"`
}

type Images struct {
	Featured Image `json:"featured"`
}

type Category struct {
	Slug string `json:"slug"`
}

type Tag struct {
	Slug string `json:"slug"`
}

type Recipe struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Excerpt    string     `json:"excerpt"`
	Slug       string     `json:"slug"`
	Images     Images     `json:"images"`
	Categories []Category `json:"categories"`
	Tags       []Tag      `json:"tags"`
}

type CategoryData struct {
	Posts []Recipe `json:"posts"`
}

type Response struct {
	PageProps struct {
		Category CategoryData `json:"category"`
	} `json:"pageProps"`
}

const baseImageURL = "https://kasviskapinastor.blob.core.windows.net/images/"

func (r *Recipe) ToPost() base.Post {
	hashtags := []string{"kasviskapina", "vegaani", "vegaaniresepti"}
	for _, tag := range r.Tags {
		hashtags = append(hashtags, tag.Slug)
	}

	imgURL := baseImageURL + r.Images.Featured.Filename

	return base.Post{
		ID:           slugToID(r.Slug),
		Title:        strings.ToUpper(r.Title),
		Description:  r.Excerpt,
		ThumbnailURL: imgURL,
		ImageURL:     imgURL,
		URL:          "https://kasviskapina.fi/reseptit/" + r.Slug,
		Added:        true,
		Hashtags:     hashtags,
	}
}

func FetchNewPosts(
	recipesFilePath string,
	httpGetter func(string, string) ([]byte, error),
	_ func(string, url.Values, string) (data []byte, err error),
	previewOnly bool,
) (base.RecipeBank, error) {
	bank, err := doFetchNewPosts(recipesFilePath, httpGetter, previewOnly)
	if err == nil {
		return bank, nil
	}

	slog.Warn("kk fetch failed, falling back to cached posts", "error", err)

	posts, _ := base.LoadExistingPosts(recipesFilePath)

	return base.RecipeBank{
		Posts:       posts,
		UsedIDsPath: UsedIDsPath,
	}, nil
}

func doFetchNewPosts(
	recipesFilePath string,
	httpGetter func(string, string) ([]byte, error),
	previewOnly bool,
) (bank base.RecipeBank, err error) {
	defer err2.Handle(&err)

	urlRes := string(try.To1(
		httpGetter("https://www.kasviskapina.fi/", ""),
	))
	endIndex := strings.Index(urlRes, "/_buildManifest.js")

	if endIndex < 0 {
		return bank, fmt.Errorf("%w: build manifest not found", errFeed)
	}

	startIndex := strings.LastIndex(urlRes[:endIndex], "/")
	hash := urlRes[startIndex+1 : endIndex]

	res := try.To1(
		httpGetter(
			fmt.Sprintf("https://www.kasviskapina.fi/_next/data/%s/fi/kategoriat/paaruoka.json?slug=paaruoka", hash),
			""),
	)

	var apiResponse Response

	try.To(json.Unmarshal(res, &apiResponse))

	posts := make([]base.Post, 0)

	for _, recipe := range apiResponse.PageProps.Category.Posts {
		for _, cat := range recipe.Categories {
			if cat.Slug == "paaruoka" {
				posts = append(posts, recipe.ToPost())

				break
			}
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID > posts[j].ID
	})

	slog.Info("Found posts for kk", "count", len(posts))

	if len(posts) == 0 {
		return bank, fmt.Errorf("%w: zero posts in response", errFeed)
	}

	if !previewOnly {
		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		try.To(encoder.Encode(posts))
		try.To(os.WriteFile(recipesFilePath, buffer.Bytes(), base.WritePerm))
	}

	return base.RecipeBank{
		Posts:       posts,
		UsedIDsPath: UsedIDsPath,
	}, nil
}
