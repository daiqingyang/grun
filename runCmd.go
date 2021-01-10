package main

import (
	"bufio"
	"errors"
	"io"
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
	cmd, e = getSafeCmd(cmd)
	if e != nil {
		logger.Print(e)
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
	}
	return
}
func copyAndRun(ip string, cmd string) {
	defer func() {
		wg.Done()
		<-concurrent
		if e := recover(); e != nil {
			logger.Println(e)
			return
		}
	}()
	direct := true
	client, e := connect(ip)
	if e != nil {
		logger.Println("client error:", e)
		return
	}
	defer client.Close()

	t := time.Now().Unix()
	fName := path.Base(cmd)
	destFile := "." + strconv.Itoa(int(t)) + "." + fName
	fullCmd := "./" + destFile + ";rm " + destFile

	if _, e := scp(client, cmd, destFile, direct); e != nil {
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
	out, _ := scp(c, srcFilePath, destFile, direct)
	mixOut(ip, out)
}

func run(ip string, cmd string) {
	defer func() {
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
	for username, password := range userPasswords {
		cConfig := &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Second * 2,
		}
		client, e = ssh.Dial("tcp", addr, cConfig)
		if e != nil {
			continue
		} else {
			break
		}
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
	//判断目标文件是否是目录，以及如果是文件，是否要备份
	if finfo, e = newClient.Stat(dest); e != nil {

		if !os.IsNotExist(e) {
			logger.Print(e)
			return
		}
	} else {
		if finfo.IsDir() {
			srcName := path.Base(local)
			dest = dest + "/" + srcName
		}

		if cfg.BackOnCopy {
			suffix := time.Now().Format("20060102150405")
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
			out, e = session.CombinedOutput(backupCmd)
			if e != nil {
				logger.Println(e)
				return
			}
		}

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
