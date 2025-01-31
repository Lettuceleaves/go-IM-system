package main

import (
	"fmt"
	"net"
	"sync"
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

func (server *Server) BroadCast(User *User, msg string) {
	server.Messege <- fmt.Sprintf("[%s]: %s", User.Name, msg)
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
	User := NewUser(Conn)
	server.maplock.Lock()
	server.UserTable[User.Name] = User
	server.maplock.Unlock()
	server.BroadCast(User, fmt.Sprintf("用户 %s 已上线", User.Name))
	select {}
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
			fmt.Println(".", err)
			continue
		}
		go server.Handle(Conn)
	}
}
