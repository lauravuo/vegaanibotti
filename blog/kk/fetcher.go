package kk

import (
	"bytes"
	"encoding/base64"
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
	"github.com/lauravuo/vegaanibotti/myhttp"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const RecipesPath = base.DataPath + "/kk/recipes.json"

const UsedIDsPath = base.DataPath + "/kk/used.json"

type Recipe struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Excerpt       string `json:"excerpt"`
	Slug          string `json:"slug"`
	FeaturedImage struct {
		Nodes struct {
			Src string `json:"sourceUrl"`
		} `json:"node"`
	} `json:"featuredImage"`
	Categories struct {
		Nodes []struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"nodes"`
	} `json:"categories"`
}

type Response struct {
	PageProps struct {
		Posts []Recipe `json:"posts"`
	} `json:"pageProps"`
}

func (r *Recipe) ToPost() base.Post {
	// ID
	strID := try.To1(base64.StdEncoding.DecodeString(r.ID))
	postID := try.To1(strconv.ParseInt(strings.Split(string(strID), ":")[1], 10, 64))

	// Title
	titleBytes := make([]byte, len([]byte(r.Title)))

	caser := cases.Title(language.Finnish)
	_, _ = try.To2(caser.Transform(titleBytes, []byte(r.Title), true))

	// Description
	startIndex := strings.Index(r.Excerpt, ">")
	endIndex := strings.LastIndex(r.Excerpt, "<")

	if endIndex < 0 {
		endIndex = len(r.Excerpt)
	}

	const baseImageURL = "https://www.kasviskapina.fi/_next/image?url=https://kasviskapinastor.blob.core.windows.net/images"

	return base.Post{
		ID:           postID,
		Title:        string(titleBytes),
		Description:  r.Excerpt[startIndex+1 : endIndex],
		ThumbnailURL: fmt.Sprintf("%s%s&w=384&q=75", baseImageURL, r.FeaturedImage.Nodes.Src),
		ImageURL:     baseImageURL + r.FeaturedImage.Nodes.Src + "&w=1920&q=75",
		URL:          "https://kasviskapina.fi/reseptit/" + r.Slug,
		Added:        true,
		Hashtags:     []string{"kasviskapina", "vegaani", "vegaaniresepti"},
	}
}

func FetchNewPosts(
	recipesFilePath string,
	_ func(string, string) ([]byte, error),
	_ func(string, url.Values, string) (data []byte, err error),
	previewOnly bool,
) (base.RecipeBank, error) {
	urlRes := string(try.To1(
		myhttp.DoGetRequest("https://www.kasviskapina.fi/", ""),
	))
	endIndex := strings.Index(urlRes, "/_buildManifest.js")
	startIndex := strings.LastIndex(urlRes[:endIndex], "/")

	// No pagination currently?, fetch all at once
	res := try.To1(
		myhttp.DoGetRequest(
			fmt.Sprintf("https://www.kasviskapina.fi/_next/data/%s/fi/kategoriat/reseptit.json?slug=paaruoka", urlRes[startIndex+1:endIndex]),
			""),
	)

	var apiResponse Response

	try.To(json.Unmarshal(res, &apiResponse))

	posts := make([]base.Post, 0)

	for _, receipt := range apiResponse.PageProps.Posts {
		for _, category := range receipt.Categories.Nodes {
			if category.Slug == "paaruoka" {
				posts = append(posts, receipt.ToPost())

				break
			}
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID > posts[j].ID
	})

	slog.Info("Found posts for kk", "count", len(posts))

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
