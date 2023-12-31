package bot

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

type FB struct {
	PageID      string
	AccessToken string
	PostJSON    func(path, method string, values interface{}, authHeader string) ([]byte, error)
}

func InitFB() *FB {
	return &FB{
		PageID:      os.Getenv("FACEBOOK_PAGE_ID"),
		AccessToken: os.Getenv("FACEBOOK_PAGE_TOKEN"),
		PostJSON:    myhttp.DoJSONRequest,
	}
}

func (f *FB) PostToFB(post *base.Post) error {
	data := make(map[string]any)
	data["message"] = post.Summary()
	data["link"] = post.URL
	data["published"] = "true"
	status := try.To1(f.PostJSON(
		fmt.Sprintf("https://graph.facebook.com/v18.0/%s/feed", f.PageID),
		"POST",
		data,
		"Bearer "+f.AccessToken),
	)
	slog.Info("post sent to facebook", "status", string(status))

	return nil
}
