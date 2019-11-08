package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const gClientID = "google-client-id"

const gTokenAPI = "https://oauth2.googleapis.com/tokeninfo"
const gAboutAPI = "https://www.googleapis.com/oauth2/v2/userinfo"

func LoginGoogle(accessToken string) (*User, error) {
	token, err := exchangeGoogleToken(accessToken)
	if err != nil {
		return nil, err
	}

	u, err := getGoogleProfile(token)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func exchangeGoogleToken(token string) (string, error) {
	r, err := http.NewRequest(http.MethodGet, gTokenAPI, nil)
	if err != nil {
		return "", err
	}

	q := r.URL.Query();
	q.Add("access_token", token)
	r.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}

	type tokenInfo struct {
		Azp           string `json:"azp"`
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Scope         string `json:"scope"`
		Exp           string `json:"exp"`
		ExpiresIn     string `json:"expires_in"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		AccessType    string `json:"access_type"`

		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	var tknInfo tokenInfo

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&tknInfo); err != nil {
		return "", err
	}

	if tknInfo.Aud == "" {
		return "", fmt.Errorf("%s: %s", tknInfo.Error, tknInfo.ErrorDescription)
	}

	if tknInfo.Aud != gClientID || tknInfo.Azp != gClientID {
		return "", fmt.Errorf("token verification failed")
	}

	return token, nil
}

func getGoogleProfile(token string) (*User, error) {
	r, err := http.NewRequest(http.MethodGet, gAboutAPI, nil)
	if err != nil {
		return nil, err
	}

	q := r.URL.Query()
	q.Add("fields", "email,id,name,picture")
	q.Add("access_token", token)
	r.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	type userInfo struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture string `json:"picture"`

		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	var ui userInfo

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&ui); err != nil {
		return nil, err
	}

	if ui.ID == "" {
		return nil, fmt.Errorf("[Code:%d] %s: %s", ui.Error.Code, ui.Error.Status, ui.Error.Message)
	}

	u := User{
		Username:    ui.Email,
		DisplayName: ui.Name,
		OauthSource: Google,
		OauthID:     ui.ID,
	}

	return &u, nil
}
