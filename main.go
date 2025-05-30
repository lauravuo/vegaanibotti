package main

import (
	"log/slog"
	"os"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/lauravuo/vegaanibotti/bot"
	"github.com/lauravuo/vegaanibotti/bot/img"
)

const (
	boldFontFile = "./bot/img/font/Amatic_SC/AmaticSC-Bold.ttf"
	regFontFile  = "./bot/img/font/Amatic_SC/AmaticSC-Regular.ttf"
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

		// Generate and upload image
		bucketURL := os.Getenv("CLOUD_BUCKET_URL")
		imageFile, smallImageFile := img.GenerateThumbnail(
			&chosenPost,
			"./bot/img/vegaanibotti.png",
			"image",
			boldFontFile,
			regFontFile,
		)
		paths := img.UploadToCloud([]string{imageFile, smallImageFile})
		chosenPost.ImageURL = bucketURL + "/" + paths[0]
		chosenPost.ThumbnailURL = bucketURL + "/" + paths[1]

		s := bot.InitSite()
		try.To(s.PostToSite(&chosenPost))

		i := bot.InitIG()
		try.To(i.PostToIG(&chosenPost))

		f := bot.InitFB()
		try.To(f.PostToFB(&chosenPost))

		x := bot.InitX()
		try.To(x.PostToX(&chosenPost))
	}
}
