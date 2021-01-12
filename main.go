package main

import (
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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
	}
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
	LogPath       string
	Debug         bool
	Forks         int
	AddNewline    bool
	RemoteRun     bool
	Copy          bool
	UserPasswords []map[string]string
	Cache         bool
	CacheFile     string
	Sshport       int
	Become        bool
	ColorPrint    bool
	BackOnCopy    bool
	PrivateKeys   []map[string]string
	AuthMethod    string
	TimeOut       time.Duration
	Alias         map[string]string
	ShortCuts     map[string]string
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

	}

	// copyAndRun("10.58.165.165", "hostname")
}
func runServer() {
	r := gin.Default()
	r.Run(":10240")
}
