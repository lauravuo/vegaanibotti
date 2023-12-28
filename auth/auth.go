package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/lauravuo/vegaanibotti/myhttp"
)

type Response struct {
	//nolint:tagliatelle
	AccessToken string `json:"access_token"`
	//nolint:tagliatelle
	RefreshToken string `json:"refresh_token"`
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
			// happy end: token parsed successfully
			os.WriteFile(
				"./data/.envrc",
				[]byte("export NEW_X_REFRESH_TOKEN="+response.RefreshToken),
				blog.WritePerm)

			return response.AccessToken
		}
	}

	return ""
}
