package blog

import (
	"log/slog"
	"net/url"
	"os"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/blog/cc"
	"github.com/lauravuo/vegaanibotti/blog/kk"
	"github.com/lauravuo/vegaanibotti/blog/vh"
	"github.com/lauravuo/vegaanibotti/blog/vmm"
	"github.com/lauravuo/vegaanibotti/blog/vv"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

type fetcher struct {
	fn func(
		string,
		func(string, string) ([]byte, error),
		func(string, url.Values, string) (data []byte, err error),
		bool,
	) (base.RecipeBank, error)
	recipesPath string
	author      string
}

func getFetchers() map[string]fetcher {
	return map[string]fetcher{
		"cc":  {cc.FetchNewPosts, cc.RecipesPath, "Chocochili"},
		"vh":  {vh.FetchNewPosts, vh.RecipesPath, "Vegaanihaaste"},
		"kk":  {kk.FetchNewPosts, kk.RecipesPath, "Kasviskapina"},
		"vmm": {vmm.FetchNewPosts, vmm.RecipesPath, "Viimeistä murua myöten"},
		"vv":  {vv.FetchNewPosts, vv.RecipesPath, "Vegeviettelys"},
	}
}

func FetchNewPosts(
	previewOnly bool,
) (base.Collection, error) {
	fetchers := getFetchers()
	collection := make(base.Collection)
	entries := try.To1(os.ReadDir(base.DataPath))

	for _, entry := range entries {
		if entry.IsDir() {
			slog.Info("Fetching recipes", "id", entry.Name())
			fetch := fetchers[entry.Name()]
			collection[entry.Name()] = try.To1(fetch.fn(
				fetch.recipesPath,
				myhttp.DoGetRequest,
				myhttp.DoPostRequest,
				previewOnly,
			))
		}
	}

	return collection, nil
}
