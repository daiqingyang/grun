package main

import (
	"log"
	"sync"
	"time"

	"grun/httpserver"
)

var (
	//des encrypt/decrypt key
	key = []byte("d9d9skrifjqlapfyrjfnamfk")

	cfgFile string = "grun.cfg"
	cfg     config = config{
		LogPath:    "go.log",
		Debug:      false,
		Cache:      true,
		CacheFile:  ".grunCache",
		Sshport:    22,
		ColorPrint: true,
		AddNewline: true,
		BackOnCopy: true,
		AuthMethod: "smart",
		TimeOut:    time.Second * 2,
		Version:    "0.3",
		DestIPList: "",
	}
	rt         runtimeConfig = runtimeConfig{}
	cmd        string
	wg         sync.WaitGroup
	concurrent chan int
	aliveList  []string
	// mutex      sync.RWMutex
	logger *log.Logger
	server *bool
	client *bool
)

type config struct {
	Version        string
	LogPath        string
	Debug          bool
	Forks          int
	Interval       int
	AddNewline     bool
	RemoteRun      bool
	Copy           bool
	UserPasswords  []map[string]string
	Cache          bool
	CacheFile      string
	Sshport        int
	Become         bool
	ColorPrint     bool
	BackOnCopy     bool
	PrivateKeys    []map[string]string
	AuthMethod     string
	TimeOut        time.Duration
	Alias          map[string]string
	ShortCuts      map[string]string
	CronAdd        bool
	CronDel        bool
	CronAnnotation string
	AdminWeb       bool
	DestIPList     string
}
type runtimeConfig struct {
	cronTmpFile string
}

func init() {
	ParseConfigFile(cfgFile)
	getCmdLineArg()
	SetLog()

}

func main() {

	// index := readCache("b.a.a.a")
	// fmt.Println(index)
	// writeCache("b.a.a.a", 20000)
	// os.Exit(1)
	if cmd != "" {
		parseAndRun(cmd)

	} else if cfg.AdminWeb {
		httpserver.RunAdminWeb()

	}

}
