package main

import (
	"log/slog"
	"os"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/lauravuo/vegaanibotti/bot"
)

func main() {
	fetchOnly := false
	if len(os.Args) > 1 {
		fetchOnly = os.Args[1] == "--fetch"
	}

	posts := try.To1(blog.FetchNewPosts(
		fetchOnly,
	))

	if !fetchOnly {
		chosenPost := blog.ChooseNextPost(posts, blog.UsedBlogsIDsPath)
		slog.Info("Chosen post",
			"title", chosenPost.Title,
			"description", chosenPost.Description,
			"url", chosenPost.URL)

		m := bot.InitMastodon()
		try.To(m.PostToMastodon(&chosenPost))

		x := bot.InitX()
		try.To(x.PostToX(&chosenPost))
	}
}
