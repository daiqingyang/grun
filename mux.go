package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/hashicorp/yamux"
)

//StartServer is server mode
func StartServer() {
	listen, e := net.Listen("tcp", ":10240")
	if e != nil {
		logger.Fatal(e)
	}
	for {
		conn, e := listen.Accept()
		if e != nil {
			logger.Println(e)
		}
		go handleServer(conn)

	}
}

func handleServer(conn net.Conn) {
	defer conn.Close()
	session, e := yamux.Client(conn, nil)
	if e != nil {
		logger.Println(e)
	}
	defer session.Close()
	stream, e := session.Open()
	if e != nil {
		logger.Println(e)
	}
	defer stream.Close()
	for {
		stream.Write([]byte("hello world!\n"))
		time.Sleep(time.Second * 1)
	}
}

//StartClient is client mode
func StartClient() {
	for {
		time.Sleep(time.Second * 1)
		conn, e := net.Dial("tcp", "10.58.165.166:10240")
		if e != nil {
			logger.Println(e)
			continue
		}
		session, e := yamux.Server(conn, nil)
		if e != nil {
			fmt.Println(e)
			continue
		}

		stream, e := session.Accept()
		if e != nil {
			logger.Println(e)
			continue
		}
		go handleClient(stream)
	}
}
func handleClient(stream net.Conn) {
	brd := bufio.NewReader(stream)
	for {
		str, e := brd.ReadString('\n')
		if e != nil {
			if e == io.EOF {
				break
			} else {
				logger.Println(e)
				break
			}
		}
		fmt.Println(":", str)
		time.Sleep(time.Second * 1)
	}
}
