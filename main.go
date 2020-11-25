package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/colorstring"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	wg         sync.WaitGroup
	concurrent chan int
	alive_list []string
	// mutex      sync.RWMutex
	no_newline *bool
)

func help() {
	fmt.Println()
	fmt.Printf("%s [-n] [-c] system_cmd|local_cmd_file\n\n", os.Args[0])
	fmt.Println("-n  will display ip and result in one line")
	fmt.Println("-c  will copy local file to remote and run")
	fmt.Println("\nput ips in EOF，one ip ,one line")
	fmt.Println("or put ips in pipeline")
	fmt.Println()
	os.Exit(1)
}
func main() {
	fmt.Println(log.Ldate, log.Ltime)
	os.Exit(1)
	copy := flag.Bool("c", false, "if copy cmd file to remote")
	no_newline = flag.Bool("n", false, "bewteen ip and result,if contains new line")

	flag.Parse()
	other := flag.Args()
	if len(other) < 1 {
		help()
	}
	cmd := other[0]
	concurrent = make(chan int, 100)
	nrd := bufio.NewReader(os.Stdin)
	for {
		str, e := nrd.ReadString('\n')
		if e != nil {
			if e == io.EOF {
				break
			} else {
				log.Fatal(e)
			}
		}
		origin_ip := strings.Trim(str, "\n ")
		ip_list := getIps(origin_ip)
		for _, ip := range ip_list {
			if isIp(ip) {
				concurrent <- 1
				wg.Add(1)
				if *copy {
					go copy_and_run(ip, cmd)
				} else {
					go run(ip, cmd)
				}
			}
		}
	}
	wg.Wait()
	// fmt.Println("end!")
}
func getIps(str string) (ips []string) {
	if strings.Contains(str, ",") {
		ips = strings.Split(str, ",")
	} else if strings.Contains(str, ";") {
		ips = strings.Split(str, ";")
	} else {
		ips = strings.Fields(str)
	}
	return
}
func isIp(str string) (is bool) {
	pattern := `^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`
	is, e := regexp.MatchString(pattern, str)
	logs("isIp error:", e)

	return
}
func connect(str string) (client *ssh.Client, e error) {
	user := "weblogic"
	password := "9tUz2LK8wymRLz3"
	addr := str + ":22"
	c_config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// Timeout:         time.Millisecond * 500,
	}
	client, e = ssh.Dial("tcp", addr, c_config)
	return
}

func scp(client *ssh.Client, local string, dest string) (e error) {
	new_client, e := sftp.NewClient(client)
	logs("new_client error:", e)
	defer new_client.Close()

	addr := client.Conn.RemoteAddr().String()

	local_file, e := os.Open(local)
	if e != nil {
		logs("open local file:", e)
		return
	}
	// fmt.Println(dest)
	dest_file, e := new_client.Create(dest)
	if e != nil {
		logs(addr+" create dest file:", e)
		return
	}
	defer dest_file.Close()

	_, e = io.Copy(dest_file, local_file)
	if e != nil {
		logs("io copy errr:", e)
		return
	}
	return
}

func copy_and_run(ip string, cmd string) {
	defer wg.Done()
	defer func() {
		<-concurrent
	}()
	client, e := connect(ip)
	if e != nil {
		logs("client error:", e)
		return
	}
	defer client.Close()
	t := time.Now().Unix()
	f_name := path.Base(cmd)
	dest_file := "." + strconv.FormatInt(t, 10) + "." + f_name

	if e := scp(client, cmd, dest_file); e != nil {
		return
	}

	session, e := client.NewSession()
	if e != nil {
		logs("new session error:", e)
		return
	}
	defer session.Close()
	out, e := session.CombinedOutput("chmod 755 " + dest_file)
	if e != nil {
		logs("run error:", e)
		return
	}

	session, e = client.NewSession()
	if e != nil {
		logs("new session error:", e)
		return
	}
	defer session.Close()
	out2, e := session.CombinedOutput("sudo " + "./" + dest_file + "&&rm " + dest_file)
	if e != nil {
		logs(ip+" run error2:", e)

		return
	}
	colorstring.Print("[red]" + ip + "\n" + "[white]" + string(out) + string(out2))

}
func run(ip string, cmd string) {
	defer wg.Done()
	defer func() {
		<-concurrent
	}()

	client, e := connect(ip)
	if e != nil {
		logs("client error:", e)
		return
	}
	defer client.Close()

	session, e := client.NewSession()
	if e != nil {
		logs("new session error:", e)
		return
	}
	defer session.Close()
	out, e := session.CombinedOutput(cmd)
	if e != nil {
		if *no_newline {
			colorstring.Print("[red]" + ip + " " + "[white]" + string(out) + e.Error() + "\n")

		} else {
			colorstring.Print("[red]" + ip + "\n" + "[white]" + string(out) + e.Error() + "\n")
		}
	} else {
		if *no_newline {
			colorstring.Print("[red]" + ip + " " + "[white]" + string(out))
		} else {
			colorstring.Print("[red]" + ip + "\n" + "[white]" + string(out))
		}
	}
}
func logs(str string, e error) {
	if e != nil {
		log.Println(str, e)
	}
}
