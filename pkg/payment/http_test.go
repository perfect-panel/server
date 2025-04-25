package payment

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHttp(t *testing.T) {
	t.Skipf("Skip TestHttp test")
	router := gin.Default()
	router.LoadHTMLGlob("./*")
	router.GET("/stripe", func(c *gin.Context) {
		c.HTML(http.StatusOK, "stripe.html", gin.H{
			"title":   "Gin HTML Example",
			"message": "Hello, Gin!",
		})
	})
	_ = router.Run(":8989")
}
