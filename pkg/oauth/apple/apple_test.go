package apple

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAppleLogin(t *testing.T) {
	t.Skipf("Skip TestAppleLogin test")
	router := gin.Default()
	router.LoadHTMLGlob("./*")
	router.GET("/apple", func(c *gin.Context) {
		c.HTML(http.StatusOK, "apple.html", gin.H{
			"title":   "Gin HTML Example",
			"message": "Hello, Gin!",
		})
	})
	router.POST("/auth/apple/callback", func(c *gin.Context) {
		var req CallbackRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}
		handleAppleCallBack(c, req)
	})
	_ = router.RunTLS(":8443", "certificate.crt", "private.key")
}

func handleAppleCallBack(ctx context.Context, request CallbackRequest) {
	fmt.Printf("request: %+v\n", request)
	// validate the token
	client, err := New(Config{
		TeamID:       TeamID,
		ClientID:     ClientID,
		KeyID:        KeyID,
		ClientSecret: ClientSecret,
		RedirectURI:  "https://test.ppanel.dev:8443/auth/apple/callback",
	})
	if err != nil {
		fmt.Println("error creating apple client: " + err.Error())
		return
	}
	resp, err := client.VerifyWebToken(ctx, request.Code)
	if err != nil {
		fmt.Println("error verifying token: " + err.Error())
		return
	}
	if resp.Error != "" {
		fmt.Printf("apple returned an error: %s - %s\n", resp.Error, resp.ErrorDescription)
		return
	}

	// Get the unique user ID
	unique, err := GetUniqueID(resp.IDToken)
	if err != nil {
		fmt.Println("error getting unique id: " + err.Error())
		return
	}
	// Get the email
	claim, err := GetClaims(resp.IDToken)
	if err != nil {
		fmt.Println("failed to get claims: " + err.Error())
		return
	}
	email := (*claim)["email"]
	emailVerified := (*claim)["email_verified"]
	isPrivateEmail := (*claim)["is_private_email"]

	// Voila!
	log.Printf("\n unique: %s \n email: %s \n email_verified: %v \n is_private_email: %v", unique, email, emailVerified, isPrivateEmail)
}
