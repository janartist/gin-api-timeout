package gin_api_timeout

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func route() *gin.Engine {
	m := NewTimeoutManager(time.Second*1, nil)
	r := gin.Default()
	r.Use(TimeoutMiddleware(m))
	r.GET("/test", func(c *gin.Context) {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done(): //取出值即说明是结束信号
					fmt.Println("收到信号，父context的协程退出,time=", time.Now().Unix())
					return
				default:
					fmt.Printf("[gotiune] run %d ...\n", time.Now().Unix())
					time.Sleep(time.Millisecond * 500)
				}

			}
		}(c.Request.Context())
		time.Sleep(time.Second * 10)
		c.String(200, "test-res")
	})
	return r
}

func performRequest(method, target string, router *gin.Engine) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

func TestTimeoutMiddleware(t *testing.T) {

	route().Run()

	w := performRequest("GET", "/test", route())

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("status code error is %d", w.Code)
	}
}
