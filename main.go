package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	Facebook = iota + 1
	Google
)

type LoginRequest struct {
	AccessToken string `json:"accessToken"`
	OauthSource int    `json:"oauthSource"`
}

type User struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	OauthSource int    `json:"oauthSource"`
	OauthID     string `json:"oauthId"`
}

func login(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var loginRequest LoginRequest
	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		log.Fatal(err)
	}

	var user *User
	if loginRequest.OauthSource == Facebook {
		user, err = LoginFacebook(loginRequest.AccessToken)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		user, err = LoginGoogle(loginRequest.AccessToken)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	err = json.NewEncoder(w).Encode(*user)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/login", login)
	err := http.ListenAndServe(":7788", nil)
	if err != nil {
		log.Fatal(err)
	}
}
