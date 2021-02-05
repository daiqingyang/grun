package httpserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func groupList(c *gin.Context) {
	var groups []Group
	config.db.Preload("Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "Username")
	}).Find(&groups)
	c.JSON(200, groups)
}
func groupAdd(c *gin.Context) {
	msg := "create"
	var group Group
	if e := c.BindJSON(&group); e != nil {
		c.JSON(200, gin.H{
			"error": e.Error(),
		})
		return
	}
	var rst Group
	config.db.First(&rst, "name=?", group.Name)

	if rst.ID != 0 {
		fmt.Println(rst)
		msg = "exists"
	} else {
		group.ID = 0
		config.db.Save(&group)
	}
	c.JSON(200, gin.H{
		"group": group,
		"msg":   msg,
	})

}
func groupUpdate(c *gin.Context) {
	var group Group
	var rst Group
	var msg = "not found"
	if e := c.ShouldBindJSON(&group); e != nil {
		c.JSON(200, gin.H{
			"error": e.Error(),
		})
		return
	}
	config.db.First(&rst, "name=?", group.Name)
	if rst.ID != 0 {
		group.ID = 0
		config.db.Model(&rst).Updates(group)
		msg = "updated"
	}
	c.JSON(200, gin.H{
		"msg": msg,
	})
}

func groupDel(c *gin.Context) {
	var group Group
	var rst Group
	if e := c.BindJSON(&group); e != nil {
		c.JSON(200, gin.H{
			"error": e.Error(),
		})
		return
	}
	config.db.First(&rst, "name=?", group.Name)
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
