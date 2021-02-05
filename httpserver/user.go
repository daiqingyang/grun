package httpserver

import (
	"crypto/md5"
	"fmt"

	"github.com/gin-gonic/gin"
)

func userList(c *gin.Context) {
	var users []UserLocal
	config.db.Preload("Groups").Find(&users)
	for index, _ := range users {
		users[index].Password = ""
	}
	c.JSON(200, users)
}
func userAdd(c *gin.Context) {
	msg := "create"
	var user UserLocal
	if e := c.BindJSON(&user); e != nil {
		c.JSON(200, gin.H{
			"error": e.Error(),
		})
		return
	}
	var rst UserLocal
	config.db.Preload("Groups").First(&rst, "username=?", user.Username)

	if rst.ID != 0 {
		fmt.Println(rst)
		msg = "exists"
		user = rst
	} else {
		user.ID = 0
		user.Password = passwdEncrypt(user.Password)
		for i, group := range user.Groups {
			if group.ID == 0 {
				config.db.Find(&group, "name=?", group.Name)
			}
			user.Groups[i] = group
		}
		config.db.Save(&user)
	}
	user.Password = ""
	c.JSON(200, gin.H{
		"user": user,
		"msg":  msg,
	})

}
func userUpdate(c *gin.Context) {
	var user UserLocalForUpdate
	var rst UserLocal
	var groupName []string
	var groups []Group
	var msg = "not found"
	if e := c.ShouldBindJSON(&user); e != nil {
		c.JSON(200, gin.H{
			"error": e.Error(),
		})
		return
	}
	config.db.Preload("Groups").First(&rst, "username=?", user.Username)
	if rst.ID != 0 {
		user.ID = 0
		if user.Password != "" {
			user.Password = passwdEncrypt(user.Password)
		}
		for _, group := range user.Groups {
			groupName = append(groupName, group.Name)
		}
		user.Groups = []Group{}
		config.db.Model(&rst).Updates(&user)
		fmt.Println(groupName)
		config.db.Where("name in ?", groupName).Find(&groups)
		config.db.Model(&rst).Association("Groups").Clear()
		config.db.Model(&rst).Association("Groups").Append(&groups)
		msg = "updated"
	}
	c.JSON(200, gin.H{
		"msg": msg,
	})
}

func userDel(c *gin.Context) {
	var user UserLocalForUpdate
	var rst UserLocal
	if e := c.BindJSON(&user); e != nil {
		c.JSON(200, gin.H{
			"error": e.Error(),
		})
		return
	}
	config.db.First(&rst, "username=?", user.Username)
	if rst.ID == 0 {
		c.JSON(200, gin.H{
			"msg": "not exsit",
		})
		return
	} else {
		config.db.Delete(&rst)
	}
	c.JSON(200, gin.H{
		"msg": "deleted",
	})
}
func passwdEncrypt(password string) (newstring string) {
	array := md5.Sum([]byte(password))
	newstring = fmt.Sprintf("%x", array)
	return
}
