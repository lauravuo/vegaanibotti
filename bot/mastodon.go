package bot

import (
	"context"
	"log/slog"
	"os"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/mattn/go-mastodon"
)

type MastodonClient interface {
	PostStatus(ctx context.Context, toot *mastodon.Toot) (*mastodon.Status, error)
}

type Mastodon struct {
	Client MastodonClient
}

func InitMastodon() *Mastodon {
	return &Mastodon{
		Client: mastodon.NewClient(&mastodon.Config{
			Server:       os.Getenv("MASTODON_SERVER"),
			ClientID:     os.Getenv("MASTODON_CLIENT_ID"),
			ClientSecret: os.Getenv("MASTODON_SECRET_KEY"),
			AccessToken:  os.Getenv("MASTODON_ACCESS_TOKEN"),
		}),
	}
}

func (m *Mastodon) PostToMastodon(post *blog.Post) error {
	status := try.To1(m.Client.PostStatus(context.Background(), &mastodon.Toot{
		Status:   post.Summary(),
		Language: "fi",
	}))
	slog.Info("post sent", "status", status.ID)

	return nil
}
