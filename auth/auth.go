package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

type Response struct {
	//nolint:tagliatelle
	AccessToken string `json:"access_token"`
	//nolint:tagliatelle
	RefreshToken string `json:"refresh_token"`
	//nolint:tagliatelle
	Type string `json:"token_type"`
	//nolint:tagliatelle
	Expires int    `json:"expires_in"`
	Scope   string `json:"scope"`
}

func FetchAccessToken(clientID, clientSecret, refreshToken, endpoint string) string {
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)))

	// authentication complete - fetch the access token
	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", refreshToken)
	data, err := myhttp.DoPostRequest(
		endpoint,
		params,
		authHeader,
	)

	if err == nil {
		response := Response{}
		if err = json.Unmarshal(data, &response); err == nil {
			slog.Info("Acquired new X tokens", "type", response.Type, "scope", response.Scope, "expires", response.Expires)
			// happy end: token parsed successfully
			try.To(os.WriteFile(
				"./data/.envrc",
				[]byte("export NEW_X_REFRESH_TOKEN=\""+response.RefreshToken+"\""),
				base.WritePerm))

			return response.AccessToken
		}
	}

	return ""
}
