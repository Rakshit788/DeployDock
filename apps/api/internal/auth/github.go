package auth

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/gin-gonic/gin"
)

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
	log.Println("👉 callback code:", code)

	clientID := "Ov23liqNFN2JXBlWmM54"
	clientSecret := "f251bc9dfbc7cea50d5a70b980aa3f0a19883c09"

	// 🔹 Step 1: Exchange code for token
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
		log.Println("❌ Token request failed:", err)
		c.JSON(500, gin.H{"error": "token request failed"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Println("👉 token response:", string(respBody))

	values, _ := url.ParseQuery(string(respBody))
	accessToken := values.Get("access_token")

	log.Println("👉 parsed token:", accessToken)

	if accessToken == "" {
		c.JSON(500, gin.H{
			"error": "failed to parse token",
			"raw":   string(respBody),
		})
		return
	}

	// 🔹 Step 2: Fetch GitHub user
	userReq, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	userReq.Header.Set("Authorization", "Bearer "+accessToken)
	userReq.Header.Set("Accept", "application/json")

	client := &http.Client{}
	userResp, err := client.Do(userReq)
	if err != nil {
		log.Println("❌ GitHub user fetch failed:", err)
		c.JSON(500, gin.H{"error": "user fetch failed"})
		return
	}
	defer userResp.Body.Close()

	userBody, _ := io.ReadAll(userResp.Body)
	log.Println("👉 GitHub raw user:", string(userBody))

	var ghUser GitHubUser
	err = json.Unmarshal(userBody, &ghUser)
	if err != nil {
		log.Println("❌ JSON parse error:", err)
		c.JSON(500, gin.H{"error": "invalid github response"})
		return
	}

	log.Println("👉 Parsed GitHub User:", ghUser.ID, ghUser.Login)

	// 🚨 safety check
	if ghUser.ID == 0 {
		c.JSON(500, gin.H{
			"error": "invalid github user",
			"raw":   string(userBody),
		})
		return
	}

	// 🔹 Step 3: Insert into DB
	log.Println("👉 Inserting user into DB...")

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
		log.Println("❌ DB INSERT ERROR:", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Println("✅ Inserted user ID:", userID)

	// 🔹 Step 4: Response
	c.JSON(200, gin.H{
		"user_id":  userID,
		"username": ghUser.Login,
	})
}
