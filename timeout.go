package gin_api_timeout

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var (
	timeoutDefault = time.Minute

	responseFuncDefault responseFunc = func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{})
	}
)

func NewTimeoutManager(timeout time.Duration, repFunc responseFunc) *TimeoutManager {
	return &TimeoutManager{timeout, repFunc}
}

type responseFunc func(ctx *gin.Context)

type TimeoutManager struct {
	timeout time.Duration
	repFunc responseFunc
}

type TimeoutOpt struct {
	timeout time.Duration
	repFunc responseFunc
}

func NewTimeoutWriter(w gin.ResponseWriter, c *gin.Context) *timeoutWriter {
	return &timeoutWriter{w, c}
}

type timeoutWriter struct {
	gin.ResponseWriter
	c *gin.Context
}

//重写底层write方法，获取response存入上下文，其余不更改
func (w *timeoutWriter) Write(data []byte) (int, error) {
	if !isTimeout(w.c) {
		return w.ResponseWriter.Write(data)
	}
	return 0, nil
}
func (w *timeoutWriter) WriteHeader(statusCode int) {
	if !isTimeout(w.c) {
		w.ResponseWriter.WriteHeader(statusCode)
	}
}
func (w *timeoutWriter) Header() http.Header {
	if !isTimeout(w.c) {
		return w.ResponseWriter.Header()
	}
	return make(http.Header)
}
func TimeoutMiddleware(m *TimeoutManager) func(c *gin.Context) {
	if m.timeout == 0 {
		m.timeout = timeoutDefault
	}
	if m.repFunc == nil {
		m.repFunc = responseFuncDefault
	}
	return func(c *gin.Context) {
		//context withTimeout
		ctx, cancel := context.WithTimeout(c, m.timeout)
		//request父协程替换
		c.Request = c.Request.WithContext(ctx)
		//replace writer
		c.Writer = NewTimeoutWriter(c.Writer, c)
		go func(ctx context.Context, c *gin.Context) {
			c.Next()
		}(ctx, c)

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				m.repFunc(c)
				writeTimeout(c)
			}
			cancel()
			return
		}
	}
}
func isTimeout(c *gin.Context) bool {
	return c.GetBool("timeout_rep")
}
func writeTimeout(c *gin.Context) {
	c.Set("timeout_rep", true)
}
