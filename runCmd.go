package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//parseAndRun  解析命令行指令，并运行
func parseAndRun(cmd string) {
	var e error
	cmd = decodeShortCuts(cmd)
	cmd = decodeAliasCmd(cmd)
	cmd, e = getSafeCmd(cmd)

	if e != nil {
		logger.Print(e)
		return
	}
	if e := preProcess(); e != nil {
		logger.Println(e)
		return
	}
	concurrent = make(chan int, cfg.Forks)
	brd := bufio.NewReader(os.Stdin)
	for {
		str, e := brd.ReadString('\n')
		if e != nil {
			if e == io.EOF {
				break
			} else {
				logger.Fatal(e)
			}
		}
		originIP := strings.Trim(str, "\n  \t")
		ipList := splitIP(originIP)
		for _, ip := range ipList {
			if isIP(ip) {
				concurrent <- 1
				wg.Add(1)
				if cfg.RemoteRun {
					go copyAndRun(ip, cmd)
				} else if cfg.Copy {
					go copyOnly(ip, cmd)
				} else if cfg.CronAdd {
					go addCrontab(ip)
				} else if cfg.CronDel {
					go delCrontab(ip)
				} else {
					go run(ip, cmd)
				}
			} else {
				logger.Println(ip, "is not ip")
			}
		}
	}
	wg.Wait()
}

//做一些预处理操作
//生成cron 临时文件
func preProcess() (e error) {
	if cmd != "" {
		if cfg.CronAdd {
			if e = crontabFormatCheck(cmd); e != nil {
				return
			}
			cmd := convertCronCmd(cmd)
			var f string
			if f, e = makeCronAddTmpFile(cmd); e != nil {
				return
			}
			rt.cronTmpFile = f
		} else if cfg.CronDel {
			if e = crontabFormatCheckForDel(cmd); e != nil {
				return
			}
			var f string
			if f, e = makeCronDelTmpFile(cmd); e != nil {
				return
			}
			rt.cronTmpFile = f
		}
	}
	return
}

//getSafeCmd 处理一些简单的危险命令
//不允许rm /
//不允许无参数的crontab，会清空cron列表
func getSafeCmd(cmd string) (newCmd string, err error) {
	err = nil
	newCmd = strings.Trim(cmd, " \n\t")
	if cmd == "/" {
		err = errors.New("cmd can not be '/'")
		return
	}
	strSlice := strings.Fields(cmd)
	if strSlice[0] == "rm" {
		for _, str := range strSlice {
			if str == "/" || str == "/*" {
				err = errors.New("[danger] cmd can not  be 'rm /' or 'rm /*'")
				return
			}
		}
	} else if len(strSlice) == 1 && strSlice[0] == "crontab" {
		err = errors.New("crontab need some args")
	}
	return
}

//decodeAliasCmd 处理、转换命令别名，别名在cfg中定义
func decodeAliasCmd(cmd string) (newCmd string) {
	cmdSlice := strings.Fields(cmd)
	cmdPosition0 := cmdSlice[0]
	var newCmdSlice []string
	var cmdPositionOther []string
	if len(cmdSlice) > 1 {
		cmdPositionOther = cmdSlice[1:]
	}
	for k, v := range cfg.Alias {
		if cmdPosition0 == k {
			cmdPosition0 = v
			break
		}
	}
	newCmdSlice = append(newCmdSlice, cmdPosition0)
	for _, other := range cmdPositionOther {
		newCmdSlice = append(newCmdSlice, other)
	}
	newCmd = strings.Join(newCmdSlice, " ")
	return
}

//decodeShortCuts  只有输入的命令完全跟定义的快捷命令一样时，才转换
func decodeShortCuts(cmd string) (newCmd string) {
	newCmd = cmd
	if cmd != "" {
		for k, v := range cfg.ShortCuts {
			if cmd == k {
				newCmd = v
				break
			}
		}
	}
	return
}

//copyAndRun 把文件拷贝到远端，并执行
//默认是拷贝到家目录下，以隐藏文件名定义
func copyAndRun(ip string, localFile string) {
	defer func() {
		wg.Done()
		<-concurrent
		if e := recover(); e != nil {
			logger.Println(e)
			return
		}
	}()
	localFileAndArgs := strings.Fields(localFile)
	localFile = localFileAndArgs[0]
	args := strings.Join(localFileAndArgs[1:], " ")
	direct := true
	client, e := connect(ip)
	if e != nil {
		logger.Println("client error:", e)
		return
	}
	defer client.Close()

	t := time.Now().Unix()
	fName := path.Base(localFile)
	destFile := "." + strconv.Itoa(int(t)) + "." + fName
	fullCmd := "./" + destFile + " " + args + ";rm " + destFile

	if _, e := scp(client, localFile, destFile, direct); e != nil {
		logger.Println(e)
		return
	}

	session, e := client.NewSession()
	if e != nil {
		logger.Println("ssh create new session error:", e)
		return
	}
	defer session.Close()
	_, e = session.CombinedOutput("chmod 755 " + destFile)
	if e != nil {
		logger.Println(e)
		return
	}
	if cfg.Become {
		fullCmd = "sudo " + fullCmd
	}
	if session, e = client.NewSession(); e != nil {
		logger.Println("ssh create new session error:", e)
		return
	}
	if cfg.Debug {
		fmt.Println("fullcmd:", fullCmd)
	}
	out, e := session.CombinedOutput(fullCmd)
	mixOut(ip, out)

	if e != nil {
		logger.Println(ip+":", e)
		return
	}
}

//copyOnly 不指定目标时，传送到/tmp/目录下
func copyOnly(ip string, cmd string) {
	defer func() {
		wg.Done()
		<-concurrent
		if err := recover(); err != nil {
			logger.Println(err)
			return
		}
	}()
	var destFile string
	var direct bool
	cmdSlice := strings.Fields(cmd)
	srcFilePath := cmdSlice[0]
	srcFileName := path.Base(srcFilePath)
	if cfg.Become {
		direct = false
	} else {
		direct = true
	}
	if len(cmdSlice) > 1 {
		destFile = cmdSlice[1]
		if strings.HasSuffix(destFile, "/") {
			destFile = destFile + srcFileName
		}
	} else {
		destFile = "/tmp/" + srcFileName

	}
	c, e := connect(ip)
	if e != nil {
		logger.Println(e)
	}
	out, err := scp(c, srcFilePath, destFile, direct)
	if err != nil {
		out = append(out, []byte(err.Error())...)
	}
	mixOut(ip, out)
}

func run(ip string, cmd string) {
	defer func() {
		timer := time.NewTimer(time.Second * time.Duration(cfg.Interval))
		<-timer.C
		wg.Done()
		<-concurrent
		if err := recover(); err != nil {
			logger.Println(err)
			return
		}
	}()
	var fullCmd string
	client, e := connect(ip)
	if e != nil {
		logger.Println(ip, ":", e)
		return
	}
	defer client.Close()

	session, e := client.NewSession()
	if e != nil {
		logger.Println("new session error:", e)
		return
	}
	defer session.Close()
	if cfg.Become {
		fullCmd = "sudo " + cmd
	} else {
		fullCmd = cmd
	}
	out, e := session.CombinedOutput(fullCmd)
	mixOut(ip, out)
	if e != nil {
		logger.Println(ip, ":", e)
		return
	}

}

func connect(str string) (client *ssh.Client, e error) {
	userPasswords := cfg.UserPasswords
	port := cfg.Sshport
	addr := str + ":" + strconv.Itoa(port)
	var content []byte
	var signer ssh.Signer
	if cfg.AuthMethod == "password" || cfg.AuthMethod == "smart" {
		for _, userPasswordsMap := range userPasswords {
			for username, password := range userPasswordsMap {
				if cfg.Debug {
					fmt.Println("ssh auth try use:", username, password)
				}
				cConfig := &ssh.ClientConfig{
					User: username,
					Auth: []ssh.AuthMethod{
						ssh.Password(password),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
					Timeout:         cfg.TimeOut,
				}
				client, e = ssh.Dial("tcp", addr, cConfig)
				if e != nil {
					if cfg.Debug {
						logger.Println(e)
					}
					continue
				} else {
					return
				}
			}
		}
	} else if cfg.AuthMethod == "sshkey" || cfg.AuthMethod == "smart" {
		for _, privateKeyMap := range cfg.PrivateKeys {
			for username, privateKey := range privateKeyMap {
				if cfg.Debug {
					fmt.Println("ssh auth try use:", username, privateKey)
				}
				content, e = ioutil.ReadFile(privateKey)
				if e != nil {
					if cfg.Debug {
						logger.Println(e)
					}
					continue
				}
				signer, e = ssh.ParsePrivateKey(content)
				cConfig := &ssh.ClientConfig{
					User: username,
					Auth: []ssh.AuthMethod{
						ssh.PublicKeys(signer),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
					Timeout:         cfg.TimeOut,
				}
				client, e = ssh.Dial("tcp", addr, cConfig)
				if e != nil {
					if cfg.Debug {
						logger.Println(e)
					}
					continue
				} else {
					return
				}
			}
		}

	} else {
		e = errors.New("ssh auth method '" + cfg.AuthMethod + "' is error")
	}
	return
}

//scp direct模式，直接scp
//非direct模式，先scp到临时目录，然后mv到目标目录
func scp(client *ssh.Client, local string, dest string, direct bool) (out []byte, e error) {
	var finfo os.FileInfo
	var backupCmd string
	var session *ssh.Session
	var destFile *sftp.File
	var tmpFile string
	var parentDir string
	addr := client.Conn.RemoteAddr().String()
	ip := strings.Split(addr, ":")[0]
	newClient, e := sftp.NewClient(client)
	if e != nil {
		logger.Print("sftp create newClient error:", e)
	}
	defer newClient.Close()

	localFile, e := os.Open(local)
	if e != nil {
		logger.Println(e)
		return
	}
	defer localFile.Close()
	// fmt.Println(dest)
	//计算出父目录和文件完整路径
	if strings.HasSuffix(dest, "/") {
		parentDir = dest
		srcName := path.Base(local)
		dest = dest + srcName
	} else {
		//如果是文件的话，剥离出父目录
		names := strings.Split(dest, "/")
		parentDir = strings.Join(names[:len(names)-1], "/")
		if parentDir == "" {
			parentDir = "./"
		}
	}
	if cfg.Debug {
		fmt.Println("parentDir", parentDir, "dst:", dest)
	}
	//保证父目录要存在
	if _, e = newClient.Stat(parentDir); e != nil {

		if os.IsNotExist(e) {
			e = newClient.MkdirAll(parentDir)
			if e != nil {
				if cfg.Debug {
					logger.Println(e)
				}
				return
			}

		} else {
			logger.Print(e)
			return
		}
	}
	//如果目标文件存在并且开启备份配置，那就备份
	//如果目录文件是目录，就报错，停止继续
	finfo, e = newClient.Stat(dest)
	if e != nil {
		if !os.IsNotExist(e) {
			logger.Println(e)
			return
		}
	} else if !finfo.IsDir() {
		if cfg.BackOnCopy {
			if cfg.Debug {
				fmt.Println("backup", dest)
			}
			suffix := time.Now().Format(".20060102150405")
			backupFile := dest + suffix
			session, e = client.NewSession()
			if e != nil {
				logger.Print(e)
				return
			}
			if cfg.Become {
				backupCmd = "sudo cp " + dest + " " + backupFile
			} else {
				backupCmd = "cp " + dest + " " + backupFile
			}
			if cfg.Debug {
				fmt.Println("backupCmd:", backupCmd)
			}
			out, e = session.CombinedOutput(backupCmd)
			if e != nil {
				logger.Println(e)
				return
			}
		}
	} else {
		e = errors.New("dest:" + dest + " exists and is directory\n")
		return
	}

	if direct {
		destFile, e = newClient.Create(dest)
		if e != nil {
			logger.Println(ip+" create dest file:", e)
			return
		}
		defer destFile.Close()
		_, e = io.Copy(destFile, localFile)
		if e != nil {
			logger.Println("io.Copy errr:", e)
			return
		}
	} else {
		tmpFile = time.Now().Format(".20060102150405")
		destFile, e = newClient.Create(tmpFile)
		if e != nil {
			logger.Println(ip+" create tmp file:", e)
			return
		}
		defer destFile.Close()

		_, e = io.Copy(destFile, localFile)
		if e != nil {
			logger.Println("io.Copy errr:", e)
			return
		}
		session, e = client.NewSession()
		if e != nil {
			logger.Print(e)
			return
		}
		mvCmd := "sudo mv " + tmpFile + " " + dest
		out, e = session.CombinedOutput(mvCmd)
		if e != nil {
			logger.Println(e)
			return
		}
	}
	return
}
