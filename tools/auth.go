package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"

	"github.com/lauravuo/vegaanibotti/myhttp"
	"github.com/pkg/browser"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

func fetchUserToken() string {
	const (
		redirectURL = "http://localhost:4321"
		xLoginURL   = "https://twitter.com/i/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s&code_challenge=challenge&code_challenge_method=plain"
	)

	var (
		clientID     = os.Getenv("X_CLIENT_ID")
		clientSecret = os.Getenv("X_CLIENT_SECRET")
		authHeader   = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)))
	)

	if clientID == "" && clientSecret == "" {
		panic(fmt.Errorf("X client ID and secret missing"))
	}

	// authorization code - received in callback
	code := ""
	// local state parameter for cross-site request forgery prevention
	state := fmt.Sprint(rand.Int())
	// scope of the access
	scope := url.QueryEscape("offline.access tweet.write tweet.read users.read")
	// loginURL
	path := fmt.Sprintf(xLoginURL, clientID, redirectURL, scope, state)

	// channel for signaling that server shutdown can be done
	messages := make(chan bool)

	// callback handler, redirect from authentication is handled here
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Query())
		// check that the state parameter matches
		if s, ok := r.URL.Query()["state"]; ok && s[0] == state {
			// code is received as query parameter
			if codes, ok := r.URL.Query()["code"]; ok && len(codes) == 1 {
				// save code and signal shutdown
				code = codes[0]
				messages <- true
			}
		}
		// redirect user's browser to spotify home page
		http.Redirect(w, r, "https://www.x.com/", http.StatusSeeOther)
	})

	// open user's browser to login page
	if err := browser.OpenURL(path); err != nil {
		panic(fmt.Errorf("failed to open browser for authentication %s", err.Error()))
	}

	server := &http.Server{Addr: ":4321"}
	// go routine for shutting down the server
	go func() {
		okToClose := <-messages
		if okToClose {
			if err := server.Shutdown(context.Background()); err != nil {
				log.Println("Failed to shutdown server", err)
			}
		}
	}()
	// start listening for callback - we don't continue until server is shut down
	log.Println(server.ListenAndServe())

	// authentication complete - fetch the access token
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("code", code)
	params.Add("code_verifier", "challenge")
	params.Add("redirect_uri", redirectURL)
	data, err := myhttp.DoPostRequest(
		"https://api.twitter.com/2/oauth2/token",
		params,
		authHeader,
	)
	if err == nil {
		response := AuthResponse{}
		fmt.Println(string(data))
		if err = json.Unmarshal(data, &response); err == nil {
			// happy end: token parsed successfully
			return response.AccessToken
		}
	}
	panic(fmt.Errorf("unable to acquire X user token"))
}

func main() {
	token := fetchUserToken()
	fmt.Println(token)
}
