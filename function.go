package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/colorstring"
)

//PathExsit ...判断路径
func PathExsit(f string) (b bool, realPath string) {
	b = false
	realPath = ""
	paths := []string{"."}
	c, e := user.Current()
	if e == nil {
		homeDir := c.HomeDir
		if homeDir != "" {
			paths = append(paths, homeDir)
		}
	}
	paths = append(paths, "/etc")
	for _, p := range paths {
		realPath = p + "/" + f
		finfo, err := os.Stat(realPath)
		if err != nil {
			if cfg.Debug {
				fmt.Println("[PathExsit] error:", e)
			}
			continue
		} else {
			isDir := finfo.IsDir()
			if isDir {
				continue
			} else {
				if cfg.Debug {
					fmt.Println("[PathExsit] use config file:", realPath)
				}
				b = true
				return b, realPath
			}
		}
	}
	return b, realPath

}

//SetLog ...
func SetLog() {
	logPath := cfg.LogPath
	f, e := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if e != nil {
		fmt.Println("[SetLog] error:", e)
	}
	if cfg.Debug {
		fmt.Println("[SetLog] use log file:", logPath)
	}
	wr := io.MultiWriter(f, os.Stdout)
	logger = log.New(wr, "", log.Ldate|log.Ltime|log.Lshortfile)

}

//ParseConfigFile
func ParseConfigFile(cfgFile string) {
	b, realPath := PathExsit(cfgFile)
	if b {
		_, e := toml.DecodeFile(realPath, &cfg)
		if e != nil {
			fmt.Println("ParseConfig error:", e)
			os.Exit(2)
		}
	} else {
		defaultConfig := `logpath="go.log"
debug=false
#并发数
forks=300
sshPort=22
#是否默认sudo到root
become=false
#拷贝文件到目标机器时，是否备份老文件
backOnCopy=true
#ssh验证方法，默认是smart模式
authMethod="smart"
#authMethod="password"
#authMethod="sshkey"
#定义私钥文件，可以添加多条，轮询模式
#[[privateKeys]]
#root="$HOME/.ssh/id_rsa"
#[[privateKeys]]
#otherUser="$HOME/.ssh/id_rsa2"
#定义ssh用户名、密码，可以添加多条，轮询模式
[[userPasswords]]
root="123456"
[[userPasswords]]
oatherUser="654321"
#定义命令别名，以左边开头的命令都会转换
[alias]
ll="ls -l"
curl="curl -s"
ping="ping -W1 -c2"
#完全匹配才转换
[shortcuts]
top="top -b -n1"
free="free -m"
ps="ps aux"
df="df -h"
net="netstat -tlnp"
`
		home := "."
		if u, e := user.Current(); e != nil {
			fmt.Println(e)
		} else {
			home = u.HomeDir
		}
		f, e := os.OpenFile(home+"/"+"grun.cfg", os.O_CREATE|os.O_WRONLY, 0600)
		if e != nil {
			fmt.Println("create config file error:", e)
			os.Exit(2)
		}
		defer f.Close()
		f.Write([]byte(defaultConfig))
		fmt.Printf(`
not found config file,create a default confile file in %s dir.
please update username and password

`, home)
		os.Exit(2)
	}
	if cfg.Debug {
		fmt.Printf("[ParseConfig] cfg struct:%+v\n", cfg)
	}
}

func readCache(ip string) (index int) {
	index = -1
	var f *os.File
	u, e := user.Current()
	if e != nil {
		logger.Println("[readCache] error", e)
		return
	}
	homeDir := u.HomeDir
	cacheFile := homeDir + "/" + cfg.CacheFile
	_, err := os.Stat(cacheFile)
	if err != nil {
		logger.Println("[readCache] error:", err)

	} else {
		f, e = os.OpenFile(cacheFile, os.O_RDONLY, 0644)
		if e != nil {
			logger.Println("[readCache] error:", e)
			return
		}
		defer f.Close()
		bufReader := bufio.NewReader(f)
		for {
			str, e := bufReader.ReadString('\n')
			if e != nil {
				break
			}
			if str == "\n" || strings.HasPrefix(str, "#") {
				continue
			}
			strSlice := strings.Fields(str)
			// fmt.Println(strSlice)
			if strSlice[0] == ip {
				indexStr := strSlice[1]
				index, _ = strconv.Atoi(indexStr)
				break
			}
		}
	}
	return
}

func writeCache(ip string, index int) {
	indexStr := strconv.Itoa(index)
	appendFile := true
	modifyFile := false
	u, e := user.Current()
	if e != nil {
		logger.Println("[writeCache] error", e)
		return
	}
	homeDir := u.HomeDir
	cacheFile := homeDir + "/" + cfg.CacheFile
	if _, err := os.Stat(cacheFile); err != nil {
		if _, e := os.Create(cacheFile); e != nil {
			logger.Println("[writeCache]err:", e)
			return
		}
	}
	content, e := ioutil.ReadFile(cacheFile)
	if e != nil {
		logger.Println("[writeCache]err:", e)
		return
	}
	linesSlice := strings.Split(string(content), "\n")
	for lineNum, line := range linesSlice {
		if !strings.HasPrefix(line, "#") {
			fields := strings.Split(line, " ")
			if len(fields) == 2 {
				if fields[0] == ip {
					appendFile = false
					if fields[1] == indexStr {
						break
					} else {
						modifyFile = true
						fields[1] = indexStr
						newline := fields[0] + " " + fields[1]
						linesSlice[lineNum] = newline
						break
					}
				}
			}
		}
	}
	if modifyFile {
		newContent := strings.Join(linesSlice, "\n")
		if e := ioutil.WriteFile(cacheFile, []byte(newContent), 0644); e != nil {
			logger.Println("[writeCache]error:", e)
		}

	}
	if cfg.Debug {
		fmt.Println("[writeCache] modifyFile:", modifyFile, "appendFile:", appendFile)
	}
	if appendFile {
		f, e := os.OpenFile(cacheFile, os.O_RDWR|os.O_APPEND, 0644)
		if e != nil {
			logger.Println(e)
			return
		}
		defer f.Close()
		newline := ip + " " + indexStr + "\n"
		f.WriteString(newline)

	}
}

func splitIP(str string) []string {
	var ips []string
	if strings.Contains(str, ",") {
		ips = strings.Split(str, ",")
	} else if strings.Contains(str, ";") {
		ips = strings.Split(str, ";")
	} else if strings.Contains(str, ":") {
		ips = strings.Split(str, ":")
	} else if strings.Contains(str, "|") {
		ips = strings.Split(str, "|")
	} else {
		ips = strings.Fields(str)
	}
	newIps := []string{}
	for _, i := range ips {
		i = strings.Trim(i, " \t")
		newIps = append(newIps, i)
	}

	return newIps
}

//isIp 粗略进行正则匹配是否为ip
func isIP(str string) (is bool) {
	pattern := `^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`
	is, e := regexp.MatchString(pattern, str)
	if e != nil {
		logger.Println(e)
	}
	return
}

func runClientOrServer() {
	if *server {
		StartServer()
	}
	if *client {
		StartClient()
	}
}

func mixOut(ip string, out []byte) {
	if cfg.ColorPrint {
		if cfg.AddNewline {
			colorstring.Print("[red]" + ip + "\n" + "[white]" + string(out))
		} else {
			colorstring.Print("[red]" + ip + " " + "[white]" + string(out))

		}
	} else {
		if cfg.AddNewline {
			fmt.Print(ip + "\n" + string(out))

		} else {
			fmt.Print(ip + " " + string(out))

		}
	}
}
