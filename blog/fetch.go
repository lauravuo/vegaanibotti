package blog

import (
	"log/slog"
	"os"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/blog/cc"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

type fetcher struct {
	fn func(
		string,
		func(string, string) ([]byte, error),
		bool,
	) (base.RecipeBank, error)
	recipesPath string
}

func FetchNewPosts(
	previewOnly bool,
) (base.Collection, error) {
	fetchers := map[string]fetcher{
		"cc": {cc.FetchNewPosts, cc.RecipesPath},
	}

	collection := make(base.Collection)
	entries := try.To1(os.ReadDir(base.DataPath))

	for _, entry := range entries {
		if entry.IsDir() {
			slog.Info("Fetching recipes", "id", entry.Name())
			fetch := fetchers[entry.Name()]
			collection[entry.Name()] = try.To1(fetch.fn(
				fetch.recipesPath,
				myhttp.DoGetRequest,
				previewOnly,
			))
		}
	}

	return collection, nil
}
