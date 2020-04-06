package session

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"net/url"
)

type GoogleApi struct {
	ClientId     string
	ClientSecret string
	RedirectUri  string
}

func (g *GoogleApi) GetIdToken(ctx context.Context, code string, userEntity *UserEntity) error {
	values := url.Values{}
	values.Add("code", code)
	values.Add("client_id", g.ClientId)
	values.Add("client_secret", g.ClientSecret)
	values.Add("redirect_uri", g.RedirectUri)
	values.Add("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", values)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	gResponse := GoogleResponse{}
	if err = json.Unmarshal(respBody, &gResponse); err != nil {
		return err
	}

	token, err := jwt.Parse(gResponse.IdToken, nil)
	if token == nil {
		return err
	}

	claims := token.Claims.(jwt.MapClaims)

	userEntity.UserId = claims["sub"].(string)
	userEntity.Name = claims["name"].(string)
	userEntity.Email = claims["email"].(string)
	userEntity.EmailVerified = claims["email_verified"].(bool)
	userEntity.Picture = claims["picture"].(string)

	return nil
}

type GoogleResponse struct {
	IdToken string `json:"id_token"`
}
