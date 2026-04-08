package auth

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/gin-gonic/gin"
)

type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

func GitHubLogin(c *gin.Context) {
	clientID := "Ov23liqNFN2JXBlWmM54"

	url := "https://github.com/login/oauth/authorize?client_id=" + clientID + "&scope=user"

	c.Redirect(http.StatusTemporaryRedirect, url)
}
func GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	log.Println("callback code:", code)
	// if error comes with login check here
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", "http://localhost:8080/auth/github/callback")

	resp, err := http.PostForm(
		"https://github.com/login/oauth/access_token",
		form,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "token request failed"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Println(" token response:", string(respBody))

	values, _ := url.ParseQuery(string(respBody))
	accessToken := values.Get("access_token")

	log.Println(" parsed token:", accessToken)

	if accessToken == "" {
		c.JSON(500, gin.H{
			"error": "failed to parse token",
			"raw":   string(respBody),
		})
		return
	}

	if accessToken == "" {
		c.JSON(500, gin.H{
			"error": "empty github token",
			"raw":   string(respBody),
		})
		return
	}

	// fetch github user
	userReq, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	userReq.Header.Set("Authorization", "Bearer "+accessToken)
	userReq.Header.Set("Accept", "application/json")

	client := &http.Client{}
	userResp, err := client.Do(userReq)
	if err != nil {
		c.JSON(500, gin.H{"error": "user fetch failed"})
		return
	}
	defer userResp.Body.Close()

	userBody, _ := io.ReadAll(userResp.Body)

	var ghUser GitHubUser
	json.Unmarshal(userBody, &ghUser)

	// 🚨 safety check
	if ghUser.ID == 0 {
		c.JSON(500, gin.H{
			"error": "invalid github user",
			"raw":   string(userBody),
		})
		return
	}

	var userID int
	err = db.Pool.QueryRow(context.Background(),
		`INSERT INTO users (github_id, username, avatar_url)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (github_id)
		 DO UPDATE SET username = EXCLUDED.username,
		 avatar_url = EXCLUDED.avatar_url
		 RETURNING id`,
		ghUser.ID, ghUser.Login, ghUser.AvatarURL,
	).Scan(&userID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"user_id":  userID,
		"username": ghUser.Login,
	})
}
