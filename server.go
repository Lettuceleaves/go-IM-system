package main

import (
	"flag"
	"fmt"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP        string
	Port      int
	UserTable map[string]*User
	maplock   sync.RWMutex
	Messege   chan string
}

func NewServer(ip string, port int) (server *Server) {
	server = &Server{
		IP:        ip,
		Port:      port,
		UserTable: make(map[string]*User),
		Messege:   make(chan string),
	}
	return
}

var limitTime int64 = 60

func init() {
	flag.Int64Var(&limitTime, "l", 60, "设置超时时间, 默认是60秒")
	flag.Parse()
}

func (server *Server) BroadCast(user *User, msg string) {
	server.Messege <- fmt.Sprintf("[%s]: %s", user.Name, msg)
}

func (server *Server) ListenMessage() {
	for {
		msg := <-server.Messege
		server.maplock.Lock()
		for _, user := range server.UserTable {
			user.C <- msg
		}
		server.maplock.Unlock()
	}
}

func (server *Server) Handle(Conn net.Conn) {
	defer Conn.Close()
	user := NewUser(Conn, server)
	user.Online()
	exit := make(chan bool)
	watchdog := make(chan bool)
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := Conn.Read(buffer)
			if n == 0 {
				user.Offline(exit)
				return
			}
			if err != nil {
				fmt.Println("Conn.Read error:", err)
				exit <- true
				return
			}
			msg := string(buffer[:n])
			user.SendMsg(msg)
			watchdog <- true
		}
	}()
	for {
		select {
		case <-watchdog:
		case <-time.After(time.Duration(limitTime) * time.Second):
			user.SendMsgToUser("您已超时，即将下线")
			user.Offline(exit)
		case <-exit:
			return
		}
	}
}

func (server *Server) Start() {
	Listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}
	defer Listener.Close()
	go server.ListenMessage()
	for {
		Conn, err := Listener.Accept()
		if err != nil {
			fmt.Println("Listener.Accept error:", err)
			continue
		}
		go server.Handle(Conn)
	}
}
