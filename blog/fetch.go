package blog

import (
	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/blog/cc"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

func FetchNewPosts(
	previewOnly bool,
) (base.Collection, error) {
	collection := make(base.Collection)
	collection["cc"] = try.To1(cc.FetchNewPosts(
		cc.RecipesPath,
		myhttp.DoGetRequest,
		previewOnly,
	))

	return collection, nil
}
