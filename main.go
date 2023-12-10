package main

import (
	"log/slog"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/lauravuo/vegaanibotti/bot"
	"github.com/lauravuo/vegaanibotti/my_http"
)

func main() {
	posts := try.To1(blog.FetchNewPosts(blog.RECIPES_PATH, my_http.DoGetRequest))
	chosenPost := blog.ChooseNextPost(posts)
	chosenPost.Hashtags = []string{"chocochili", "vegaani", "vegaaniresepti"}
	slog.Info("Chosen post",
		"title", chosenPost.Title,
		"description", chosenPost.Description,
		"url", chosenPost.URL)
	try.To(bot.PostToMastodon(chosenPost))
}
