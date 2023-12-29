package blog

import (
	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/blog/cc"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

func FetchNewPosts(
	previewOnly bool,
) ([]base.Post, error) {
	return cc.FetchNewPosts(RecipesPath,
		myhttp.DoGetRequest,
		previewOnly,
	)
}
