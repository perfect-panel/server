package payment

import (
	"net/http"
	"testing"

	"github.com/perfect-panel/server/pkg/hertzx"
)

func TestHttp(t *testing.T) {
	t.Skipf("Skip TestHttp test")
	router := hertzx.Default()
	router.LoadHTMLGlob("./*")
	router.GET("/stripe", func(c *hertzx.Context) {
		c.HTML(http.StatusOK, "stripe.html", hertzx.H{
			"title":   "Hertz HTML Example",
			"message": "Hello, Hertz!",
		})
	})
	_ = router.Run(":8989")
}
