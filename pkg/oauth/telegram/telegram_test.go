package telegram

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOAuth(t *testing.T) {
	t.Skipf("Skip TestOAuth test")
	router := gin.Default()
	router.LoadHTMLGlob("./*")
	router.GET("/telegram", func(c *gin.Context) {
		c.HTML(http.StatusOK, "telegram.html", gin.H{
			"title":   "Gin HTML Example",
			"message": "Hello, Gin!",
		})
	})
	router.GET("/auth/telegram/callback", func(c *gin.Context) {

	})
	_ = router.RunTLS(":443", "server.crt", "server.key")
}

func TestBase64(t *testing.T) {
	text := "eyJpZCI6ODI0NjI2ODAzLCJmaXJzdF9uYW1lIjoiQ2hhbmcgbHVlIiwibGFzdF9uYW1lIjoiVHNlbiIsInVzZXJuYW1lIjoidGVuc2lvbl9jIiwicGhvdG9fdXJsIjoiaHR0cHM6XC9cL3QubWVcL2lcL3VzZXJwaWNcLzMyMFwvYU1LNkhEc0pqc2V1YldRYmt2NGlYOHZCRUF6N0hWU3g3dkFuRDBLZ0tFVS5qcGciLCJhdXRoX2RhdGUiOjE3Mzc4MTkwNzQsImhhc2giOiI5M2I1ZDg3Zjc3NjE2YjBjMTM0OTAxYmYwMDg3MTc4YjJiYmZlYzA1MTlkMWVmMDJhZjFjMGNlOTAzM2ZiNGFlIn0"
	var token = "7651491571:AAEVQma6niHhtqEYDowAEpPo6Fq69BWvRU8"

	data, err := ParseAndValidateBase64([]byte(text), token)
	if err != nil {
		t.Error(err)
	}
	t.Log(*data.Id)

}
