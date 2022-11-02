## gin timeout Middleware

- api业务逻辑默认1分钟超时，可通过 c.Request.Context() 上下文通知cancel各个 IO客户端连接


### Quick start

```shell
go get github.com/janartist/gin-api-timeout
```

```go
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
```