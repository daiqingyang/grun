package httpserver

import (
	"fmt"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func MyFormatterLog() gin.HandlerFunc {
	return gin.LoggerWithFormatter(
		func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[GIN] %s %s %s %s %s %d %s \"%s\" %s %s\n",
				param.ClientIP,
				param.TimeStamp.Format("20060102 150405"),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.Request.Host,
				param.ErrorMessage,
			)
		},
	)
}
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

//session midddleware
func MySession() gin.HandlerFunc {
	store := sessions.NewCookieStore([]byte("iam_gin"))
	store.Options(sessions.Options{
		MaxAge: 60 * 60 * 24 * 30,
	})
	return sessions.Sessions("mySessions", store)
}
