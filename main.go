package main

import (
	"log/slog"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/chochidon/blog"
	"github.com/lauravuo/chochidon/my_http"
)

func main() {
	posts := try.To1(blog.FetchNewPosts(blog.RECIPES_PATH, my_http.DoGetRequest))
	chosenPost := blog.ChooseNextPost(posts)
	slog.Info("Chosen post",
		"title", chosenPost.Title,
		"description", chosenPost.Description,
		"url", chosenPost.URL)
}
