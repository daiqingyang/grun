package httpserver

import (
	"fmt"
	"io"
	"os"

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
func MyLog() gin.HandlerFunc {
	var out *os.File
	var err error
	//set log format
	var Formatter gin.LogFormatter = func(param gin.LogFormatterParams) string {
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
	}
	//set output file
	fName := "gin.log"
	out, err = os.OpenFile(fName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
		out = os.Stdout
	}
	var Output io.Writer = out
	config := gin.LoggerConfig{
		Formatter: Formatter,
		Output:    Output,
	}

	return gin.LoggerWithConfig(config)
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
