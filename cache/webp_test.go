package cache

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWebpGet(t *testing.T) {

	r := gin.Default()

	cachefile := r.Group("/cachefile")
	{
		cachefile.GET("/webp", func(ctx *gin.Context) {
			b, err := Webp.Get(ctx.Query("key"))
			if err != nil {
				ctx.String(404, "Not Found")
				return
			}
			ctx.Data(200, "image/webp", b)

		})
	}

	req, err := http.NewRequest("GET", "/cachefile/webp?key=20230426", nil)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	file, err := os.Create("test.webp")
	if err != nil {
		t.Error(err)
	}

	_, err = io.Copy(file, w.Body)
	if err != nil {
		t.Error(err)
	}

	file.Close()

}
