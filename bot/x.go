package bot

import (
	"log/slog"
	"os"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/auth"
	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

type X struct {
	AccessToken string
	PostJSON    func(path, method string, values interface{}, authHeader string) ([]byte, error)
}

func InitX() *X {
	return &X{
		AccessToken: auth.FetchAccessToken(
			os.Getenv("X_CLIENT_ID"),
			os.Getenv("X_CLIENT_SECRET"),
			os.Getenv("X_REFRESH_TOKEN"),
			"https://api.twitter.com/2/oauth2/token",
		),
		PostJSON: myhttp.DoJSONRequest,
	}
}

func (x *X) PostToX(post *blog.Post) error {
	data := make(map[string]any)
	data["text"] = post.ShortSummary()
	status := try.To1(x.PostJSON(
		"https://api.twitter.com/2/tweets",
		"POST",
		data,
		"Bearer "+x.AccessToken),
	)
	slog.Info("post sent to x", "status", string(status))

	return nil
}
