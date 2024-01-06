package bot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

const goToBioText = "Klikkaa reseptiin bion linkistä!"

type IG struct {
	UserID      string
	AccessToken string
	Post        func(string, url.Values, string) ([]byte, error)
}

type MediaResponse struct {
	ID string `json:"id"`
}

func InitIG() *IG {
	return &IG{
		UserID:      os.Getenv("INSTAGRAM_BUSINESS_ACCOUNT"),
		AccessToken: os.Getenv("FACEBOOK_PAGE_TOKEN"),
		Post:        myhttp.DoPostRequest,
	}
}

func (i *IG) PostToIG(post *base.Post) error {
	// add click to bio link before hashtags
	caption := post.Summary()
	index := strings.LastIndex(caption, "\n#")
	caption = caption[:index] + "\n" + goToBioText + "\n" + caption[index:]

	params := url.Values{}

	if post.ImageURL != "" {
		params.Add("image_url", post.ImageURL)
	} else {
		params.Add("image_url", post.ThumbnailURL)
	}

	params.Add("caption", caption)
	params.Add("access_token", i.AccessToken)

	res := try.To1(i.Post(
		fmt.Sprintf("https://graph.facebook.com/v18.0/%s/media", i.UserID),
		params,
		""),
	)

	var resID MediaResponse

	try.To(json.Unmarshal(res, &resID))

	params.Add("creation_id", resID.ID)
	status := try.To1(i.Post(
		fmt.Sprintf("https://graph.facebook.com/v18.0/%s/media_publish", i.UserID),
		params,
		""),
	)
	slog.Info("post sent to instagram", "status", string(status))

	return nil
}
