package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/gin-gonic/gin"
)

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubAuthRequest struct {
	Code string `json:"code"`
}

type AuthResponse struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func githubCredentials() (string, string, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", "", errors.New("GitHub OAuth is not configured")
	}
	return clientID, clientSecret, nil
}

func GitHubLogin(c *gin.Context) {
	clientID, _, err := githubCredentials()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth is not configured"})
		return
	}

	url := "https://github.com/login/oauth/authorize?client_id=" + clientID + "&redirect_uri=http://localhost:3000/auth/callback&scope=user:email%20repo"

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	clientID, clientSecret, err := githubCredentials()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth is not configured"})
		return
	}

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
		log.Println("Token request failed:", err)
		c.JSON(500, gin.H{"error": "token request failed"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	values, _ := url.ParseQuery(string(respBody))
	accessToken := values.Get("access_token")

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
		log.Println(" GitHub user fetch failed:", err)
		c.JSON(500, gin.H{"error": "user fetch failed"})
		return
	}
	defer userResp.Body.Close()

	userBody, _ := io.ReadAll(userResp.Body)

	var ghUser GitHubUser
	err = json.Unmarshal(userBody, &ghUser)
	if err != nil {
		log.Println(" JSON parse error:", err)
		c.JSON(500, gin.H{"error": "invalid github response"})
		return
	}

	//  safety check
	if ghUser.ID == 0 {
		c.JSON(500, gin.H{
			"error": "invalid github user",
			"raw":   string(userBody),
		})
		return
	}

	//  Step 3: Insert into DB
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
		log.Println(" DB INSERT ERROR:", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	//  Step 4: Generate JWT and Response
	token, err := GenerateJWT(userID, ghUser.Login)
	if err != nil {
		log.Println(" JWT generation error:", err)
		c.JSON(500, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(200, gin.H{
		"user_id":  userID,
		"username": ghUser.Login,
		"token":    token,
	})
}

// API endpoint for frontend to exchange code for user info
func GitHubAuthAPI(c *gin.Context) {
	var req GitHubAuthRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	code := req.Code
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	clientID, clientSecret, err := githubCredentials()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth is not configured"})
		return
	}

	// 🔹 Step 1: Exchange code for token
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", "http://localhost:3000/auth/callback")

	resp, err := http.PostForm(
		"https://github.com/login/oauth/access_token",
		form,
	)
	if err != nil {
		log.Println("Token request failed:", err)
		c.JSON(500, gin.H{"error": "token request failed"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	values, _ := url.ParseQuery(string(respBody))
	accessToken := values.Get("access_token")

	if accessToken == "" {
		c.JSON(500, gin.H{"error": "failed to parse token"})
		return
	}

	// 🔹 Step 2: Fetch GitHub user
	userReq, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	userReq.Header.Set("Authorization", "Bearer "+accessToken)
	userReq.Header.Set("Accept", "application/json")

	client := &http.Client{}
	userResp, err := client.Do(userReq)
	if err != nil {
		log.Println(" GitHub user fetch failed:", err)
		c.JSON(500, gin.H{"error": "user fetch failed"})
		return
	}
	defer userResp.Body.Close()

	userBody, _ := io.ReadAll(userResp.Body)
	var ghUser GitHubUser
	err = json.Unmarshal(userBody, &ghUser)
	if err != nil {
		log.Println(" JSON parse error:", err)
		c.JSON(500, gin.H{"error": "invalid github response"})
		return
	}

	if ghUser.ID == 0 {
		c.JSON(500, gin.H{"error": "invalid github user"})
		return
	}

	// 🔹 Step 3: Insert into DB
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
		log.Println(" DB INSERT ERROR:", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 🔹 Step 4: Generate JWT and return response
	jwtToken, err := GenerateJWT(userID, ghUser.Login)
	if err != nil {
		log.Println(" JWT generation error:", err)
		c.JSON(500, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(200, gin.H{
		"user_id":      userID,
		"username":     ghUser.Login,
		"token":        jwtToken,
		"github_token": accessToken,
	})
}

// GetUser returns current user info
func GetUser(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	username, _ := c.Get("username")
	c.JSON(200, gin.H{
		"id":       userID,
		"username": username,
		"login":    username, // satisfy frontend expects user.login
	})
}

