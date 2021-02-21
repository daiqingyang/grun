package httpserver

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type httpConfig struct {
	port string
}

func RunAdminWeb() {
	initDB()
	htConfig := httpConfig{
		port: "10240",
	}
	eng := gin.New()
	eng.Use(gin.Recovery(), MyLog(), CORS(), MySession())
	rootGroup := eng.Group("/")
	{
		rootGroup.GET("/", root)
		rootGroup.GET("/ping", ping)
		rootGroup.GET("/test", test)
		rootGroup.StaticFS("/static", http.Dir("static"))
	}
	//need login auth route group
	auth := eng.Group("/p")
	{
		auth.POST("/login", login)
		auth.GET("/logout", logout)
		auth.GET("/user", userList)
		auth.POST("/user", userAdd)
		auth.DELETE("/user", userDel)
		auth.PUT("/user", userUpdate)
		auth.GET("/group", groupList)
		auth.POST("/group", groupAdd)
		auth.DELETE("/group", groupDel)
		auth.PUT("/group", groupUpdate)
		auth.GET("/sync", syncTable)
	}

	eng.NoRoute(noRoute)
	eng.Run(":" + htConfig.port)

}
func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"msg": "pong",
	})
}
func syncTable(c *gin.Context) {
	config.openDebug()
	if e := config.syncDB(); e != nil {
		c.String(200, e.Error())
		return
	}
	c.String(200, "ok")
}
func login(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	loginError := true
	user := UserLocal{}
	err := c.ShouldBind(&user)
	if err == nil {
		if user.Username == "admin123" && user.Password == "123456" {
			loginError = false
			session.Set("username", user.Username)
			session.Set("logined", true)
			session.Save()
			logined := session.Get("logined")
			username := session.Get("username")
			fmt.Println(username, logined)
		}
	} else {
		c.JSON(200, gin.H{
			"status": 401,
			"error":  err.Error(),
		})
		return
	}

	if loginError {
		c.JSON(200, gin.H{
			"status": 401,
			"error":  "login error",
		})
	} else {
		c.JSON(200, gin.H{
			"status": 200,
			"msg":    "logined",
		})
	}

}
func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Set("logined", false)
	session.Save()
	c.JSON(200, gin.H{
		"status": 200,
		"msg":    "logout",
	})
}
func root(c *gin.Context) {
	session := sessions.Default(c)
	logined := session.Get("logined")
	username := session.Get("username")
	fmt.Println(username, logined)
	if logined == true {
		c.JSON(200, gin.H{
			"status":   200,
			"msg":      "welcome",
			"username": username,
		})
	} else {
		c.JSON(200, gin.H{
			"status": 401,
			"msg":    "need login",
		})
	}
}
func test(c *gin.Context) {
	var fruits = []string{"apple", "banna", "watermelon"}
	c.HTML(200, "list.html", fruits)

}
func noRoute(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "404",
	})
}
