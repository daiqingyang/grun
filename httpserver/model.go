package httpserver

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBConfig struct {
	dsn string
	db  *gorm.DB
}
type UserLocal struct {
	gorm.Model
	Username string `binding:"required,alphanum,min=6,max=12" `
	Realname string
	Password string `json:"Password,omitempty" binding:"required,min=6,max=12"`
	Phone    string
	Email    string
	Memo     string
	Title    string
	Disabled bool
	Groups   []Group `gorm:"many2many:user_groups;constraint:OnDelete:CASCADE;"`
}

//UserLocalForUpdate 当客户端进行更新\删除请求时候，去除password required
type UserLocalForUpdate struct {
	gorm.Model
	Username string `binding:"required,alphanum,min=6,max=12" `
	Realname string
	Password string
	Phone    string
	Email    string
	Memo     string
	Title    string
	Disabled bool
	Groups   []Group
}

type Group struct {
	gorm.Model
	Name  string `binding:"required"`
	Memo  string
	Users []UserLocal `gorm:"many2many:user_groups;"`
}

var (
	config *DBConfig
	db     *gorm.DB
)

func init() {

	config = &DBConfig{
		dsn: "root:daidai141@tcp(127.0.0.1:3306)/goweb?charset=utf8&parseTime=true&loc=Local",
	}
	if e := config.conncetDB(); e != nil {
		panic(e)
	}
}
func (config *DBConfig) openDebug() {
	config.db = config.db.Debug()
}
func (config *DBConfig) conncetDB() (e error) {
	config.db, e = gorm.Open(mysql.Open(config.dsn), &gorm.Config{})
	return
}
func (config *DBConfig) syncDB() (e error) {
	e = config.db.AutoMigrate(&UserLocal{}, &Group{})
	return
}
