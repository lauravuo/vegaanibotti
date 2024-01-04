package vh

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

const RecipesPath = base.DataPath + "/vh/recipes.json"

const UsedIDsPath = base.DataPath + "/vh/used.json"

type Recipe struct {
	//nolint:tagliatelle
	ID string `json:"_id"`
	//nolint:tagliatelle
	Source struct {
		//nolint:tagliatelle
		ImageURL []string `json:"image_url"`
		ID       []int    `json:"nid"`
		Title    []string `json:"title"`
		URL      []string `json:"url"`
	} `json:"_source"`
}

func (v *Recipe) ToPost() base.Post {
	url := strings.Replace(v.Source.URL[0], "http://users.", "https://", 1)
	url = strings.Replace(url, "https://users.", "https://", 1)

	return base.Post{
		ID:       int64(v.Source.ID[0]),
		Title:    v.Source.Title[0],
		URL:      url,
		ImageURL: v.Source.ImageURL[0],
		Hashtags: []string{"vegaanihaaste", "vegaani", "vegaaniresepti"},
	}
}

type SearchResponse struct {
	Responses []struct {
		Status int `json:"status"`
		Hits   struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []Recipe `json:"hits"`
		} `json:"hits"`
	} `json:"responses"`
}

func doSearch(count int, payload string) (searchRes SearchResponse, err error) {
	defer err2.Handle(&err)

	url := "https://vc-search.anima.dk/vc_fi_recipes/_msearch?"
	slog.Info("Fetching URL", "url", url, "count", count)
	res := try.To1(
		myhttp.DoJSONBytesRequest(url,
			http.MethodPost,
			[]byte(fmt.Sprintf(payload, count)),
			""),
	)

	try.To(json.Unmarshal(res, &searchRes))

	return searchRes, err
}

func FetchNewPosts(
	recipesFilePath string,
	_ func(string, string) ([]byte, error),
	previewOnly bool,
) (base.RecipeBank, error) {
	posts, maxID := base.LoadExistingPosts(recipesFilePath)

	payload := try.To1(os.ReadFile("./blog/vh/payload.json.txt"))
	count := 0
	status := http.StatusOK

	var postToAdd base.Post

	for status == http.StatusOK {
		searchRes := try.To1(doSearch(count, string(payload)))

		status = searchRes.Responses[0].Status
		count += len(searchRes.Responses[0].Hits.Hits)

		// Do not include unpublished items
		trimIndex := -1

		for index, hit := range searchRes.Responses[0].Hits.Hits {
			if strings.Contains(hit.Source.ImageURL[0], "blurrattu") {
				trimIndex = index
			}
		}

		hits := searchRes.Responses[0].Hits.Hits
		if trimIndex >= 0 {
			hits = hits[trimIndex+1:]
		}

		for _, hit := range hits {
			postToAdd = hit.ToPost()
			if postToAdd.ID > maxID {
				posts = append(posts, postToAdd)
			}
		}

		if count >= searchRes.Responses[0].Hits.Total.Value || maxID >= int64(hits[0].Source.ID[0]) {
			break
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID > posts[j].ID
	})

	slog.Info("Found posts for vh", "count", len(posts))

	if !previewOnly {
		try.To(os.WriteFile(recipesFilePath, try.To1(json.Marshal(posts)), base.WritePerm))
	}

	return base.RecipeBank{
		Posts:       posts,
		UsedIDsPath: UsedIDsPath,
	}, nil
}
