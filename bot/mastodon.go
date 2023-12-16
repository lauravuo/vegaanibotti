package bot

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/mattn/go-mastodon"
)

func PostToMastodon(post *blog.Post) error {
	client := mastodon.NewClient(&mastodon.Config{
		Server:       os.Getenv("MASTODON_SERVER"),
		ClientID:     os.Getenv("MASTODON_CLIENT_ID"),
		ClientSecret: os.Getenv("MASTODON_SECRET_KEY"),
		AccessToken:  os.Getenv("MASTODON_ACCESS_TOKEN"),
	})
	status := try.To1(client.PostStatus(context.Background(), &mastodon.Toot{
		Status: post.Title + "\n\n" +
			post.Description + "\n\n" +
			post.URL + "\n\n" +
			"#" + strings.Join(post.Hashtags, " #"),
		Language: "fi",
	}))
	slog.Info("post sent", "status", status.ID)

	return nil
}
