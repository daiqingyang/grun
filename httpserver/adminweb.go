package httpserver

import (
	"fmt"
	"io"
	"os"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type httpConfig struct {
	port string
}

func setLog() {
	logFile, e := os.OpenFile("gin.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if e != nil {
		fmt.Println(e)
	}
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)
	gin.DefaultErrorWriter = io.MultiWriter(os.Stdout, logFile)
}

func RunAdminWeb() {

	initDB()
	htConfig := httpConfig{
		port: "10240",
	}
	setLog()
	eng := gin.Default()
	eng.Use(CORS())
	//session midddleware
	store := sessions.NewCookieStore([]byte("iam_gin"))
	store.Options(sessions.Options{
		MaxAge: 60 * 60 * 24 * 30,
	})
	eng.Use(sessions.Sessions("mySessions", store))

	eng.Static("/static", "static")
	eng.LoadHTMLGlob("templates/*")
	eng.GET("/", root)
	eng.POST("/login", login)
	eng.GET("/logout", logout)
	eng.GET("/user", userList)
	eng.POST("/user", userAdd)
	eng.DELETE("/user", userDel)
	eng.PUT("/user", userUpdate)
	eng.GET("/group", groupList)
	eng.POST("/group", groupAdd)
	eng.DELETE("/group", groupDel)
	eng.PUT("/group", groupUpdate)

	eng.GET("/sync", syncTable)
	eng.GET("/ping", ping)
	eng.GET("/test", test)
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
