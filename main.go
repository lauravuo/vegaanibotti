package main

import (
	"log/slog"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/lauravuo/vegaanibotti/bot"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

func main() {
	posts := try.To1(blog.FetchNewPosts(blog.RecipesPath, myhttp.DoGetRequest))
	chosenPost := blog.ChooseNextPost(posts, blog.UsedIDsPath)
	chosenPost.Hashtags = []string{"chocochili", "vegaani", "vegaaniresepti"}
	slog.Info("Chosen post",
		"title", chosenPost.Title,
		"description", chosenPost.Description,
		"url", chosenPost.URL)

	m := bot.InitMastodon()
	try.To(m.PostToMastodon(&chosenPost))
}
