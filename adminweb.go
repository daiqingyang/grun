package main

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type httpConfig struct {
	port string
}

func runAdminWeb() {
	htConfig := httpConfig{
		port: "10240",
	}
	eng := gin.Default()
	store := sessions.NewCookieStore([]byte("iam_gin"))

	eng.Use(sessions.Sessions("mySessions", store))
	eng.Static("/static", "static")
	eng.LoadHTMLGlob("templates/*")
	eng.GET("/", root)
	eng.GET("/login", login)
	eng.POST("/login", login)
	eng.GET("/test", func(c *gin.Context) {
		var fruits = []string{"apple", "banna", "watermelon"}
		c.HTML(200, "list.html", fruits)
	})
	eng.NoRoute(func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "404",
		})
	})

	eng.Run(":" + htConfig.port)

}

func login(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	if c.Request.Method == "GET" {
		c.HTML(200, "login.html", nil)

	} else {
		error := true
		user := c.PostForm("username")
		pw := c.PostForm("pw")
		if user == "admin" && pw == "123456" {
			error = false
			session := sessions.Default(c)
			session.Set("username", user)
			session.Set("logined", true)
			session.Save()
		}
		if error {
			c.HTML(200, "login.html", error)
		} else {
			c.Redirect(302, "/")

		}

	}

}
func root(c *gin.Context) {
	session := sessions.Default(c)
	logined := session.Get("logined")
	username := session.Get("username")
	if logined == true {
		c.String(200, "welcome %s", username)
	} else {
		c.Redirect(302, "/login")
	}
}
