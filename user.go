package main

import (
	"net"
)

type User struct {
	Name string
	Add  string
	C    chan string
	conn net.Conn
}

func NewUser(conn net.Conn) (user *User) {
	UserAddr := conn.RemoteAddr().String()
	user = &User{
		Name: UserAddr,
		Add:  UserAddr,
		C:    make(chan string),
		conn: conn,
	}
	go user.ListenMessage()
	return
}

func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}
