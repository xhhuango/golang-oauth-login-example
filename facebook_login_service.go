package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const fbClientID = "fb-client-id"
const fbClientSecret = "fb-client-secret"

const fbGraphAPIBase = "https://graph.facebook.com/v4.0"
const fbTokenAPI = fbGraphAPIBase + "/oauth/access_token"
const fbMeAPI = fbGraphAPIBase + "/me"

func LoginFacebook(accessToken string) (*User, error) {
	token, err := exchangeFacebookToken(accessToken)
	if err != nil {
		return nil, err
	}

	u, err := getFacebookProfile(token)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func exchangeFacebookToken(token string) (string, error) {
	r, err := http.NewRequest(http.MethodGet, fbTokenAPI, nil)
	if err != nil {
		return "", err
	}

	q := r.URL.Query()
	q.Add("client_id", fbClientID)
	q.Add("client_secret", fbClientSecret)
	q.Add("grant_type", "fb_exchange_token")
	q.Add("fb_exchange_token", token)
	r.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}

	type exchangeToken struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   uint64 `json:"expires_in"`

		Error struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			FBTraceID string `json:"fbtrace_id"`
		} `json:"error"`
	}
	var exTkn exchangeToken

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&exTkn); err != nil {
		return "", err
	}

	if exTkn.AccessToken == "" {
		return "", fmt.Errorf("[Code:%d] %s", exTkn.Error.Code, exTkn.Error.Message)
	}

	return exTkn.AccessToken, nil
}

func getFacebookProfile(token string) (*User, error) {
	r, err := http.NewRequest(http.MethodGet, fbMeAPI, nil)
	if err != nil {
		return nil, err
	}

	q := r.URL.Query()
	q.Add("fields", "email,name,picture.width(200).height(200)")
	q.Add("access_token", token)
	r.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	type me struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				Height       uint   `json:"height"`
				Width        uint   `json:"width"`
				URL          string `json:"url"`
				IsSilhouette bool   `json:"is_silhouette"`
			} `json:"data"`
		} `json:"picture"`

		Error struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			FBTraceID string `json:"fbtrace_id"`
		} `json:"error"`
	}
	var m me

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}

	if m.ID == "" {
		return nil, fmt.Errorf("[Code:%d] %s", m.Error.Code, m.Error.Message)
	}

	u := User{
		Username:    m.Email,
		DisplayName: m.Name,
		OauthSource: Facebook,
		OauthID:     m.ID,
	}

	return &u, nil
}
