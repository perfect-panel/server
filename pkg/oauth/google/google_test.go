package google

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"

	"golang.org/x/oauth2"
)

func TestGoogleOAuth(t *testing.T) {
	t.Skipf("Skip TestGoogleOAuth test")
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/auth", handleCallback)
	http.HandleFunc("/user", handleAuth)

	fmt.Println("Server is running on http://localhost:3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	html := `<html>
		<body>
			<a href="/login">Log in with Google</a>
		</body>
	</html>`
	fmt.Fprint(w, html)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	oauthConfig := New(&Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "http://localhost:3001/auth",
	})
	url := oauthConfig.AuthCodeURL("randomstate", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != "randomstate" {
		http.Error(w, "State is invalid", http.StatusBadRequest)
		return
	}

	log.Printf("url: %v", r.URL)

	oauthConfig := New(&Config{
		ClientID:     "",
		ClientSecret: "Key",
		RedirectURL:  "http://localhost:3001/auth",
	})
	code := r.FormValue("code")
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/user?token="+token.AccessToken, http.StatusTemporaryRedirect)
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	client := New(&Config{
		ClientID:     "Id",
		ClientSecret: "Key",
		RedirectURL:  "http://localhost:3001/auth",
	})
	userInfo, err := client.GetUserInfo(token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Hello, %s", userInfo.Name)
}
