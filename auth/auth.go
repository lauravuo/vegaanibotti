package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/lauravuo/vegaanibotti/myhttp"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
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
		response := AuthResponse{}
		//fmt.Println(string(data))
		if err = json.Unmarshal(data, &response); err == nil {
			// happy end: token parsed successfully
			return response.AccessToken
		}
	}
	return ""
}
