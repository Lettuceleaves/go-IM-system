package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string
	Add    string
	C      chan string
	conn   net.Conn
	server *Server
}

func (user *User) Online() {
	user.server.maplock.Lock()
	user.server.UserTable[user.Name] = user
	user.server.maplock.Unlock()
	user.server.BroadCast(user, fmt.Sprintf("用户 %s 已上线", user.Name))
}

func (user *User) Offline(exit chan bool) {
	user.server.BroadCast(user, fmt.Sprintf("用户 %s 已下线", user.Name))
	user.server.maplock.Lock()
	delete(user.server.UserTable, user.Name)
	user.server.maplock.Unlock()
	exit <- true
}

func (user *User) SendMsgToUser(msg string) {
	user.C <- msg
}

func (user *User) UserRename(name string) {
	if _, ok := user.server.UserTable[name]; ok {
		user.SendMsgToUser("用户名已存在，请重新输入")
		return
	} else {
		user.server.maplock.Lock()
		delete(user.server.UserTable, user.Name)
		user.Name = name
		user.server.UserTable[user.Name] = user
		user.server.maplock.Unlock()
		user.SendMsgToUser(fmt.Sprintf("您的用户名已修改为 %s", name))
	}
}

func (user *User) SendMsg(msg string) {
	submsgs := strings.Split(msg, " ")
	if submsgs[0] == "用户列表" && len(submsgs) == 1 {
		user.server.maplock.RLock()
		for _, target := range user.server.UserTable {
			user.SendMsgToUser(fmt.Sprintf("[%s]:%s 在线", target.Name, target.Add))
		}
		user.server.maplock.RUnlock()
		return
	} else if submsgs[0] == "改名" && len(submsgs) == 2 {
		user.UserRename(submsgs[1])
	} else if submsgs[0] == "私聊" && len(submsgs) >= 3 {
		if _, ok := user.server.UserTable[submsgs[1]]; ok {
			target := user.server.UserTable[submsgs[1]]
			target.SendMsgToUser(fmt.Sprintf("私聊[%s]:%s", user.Name, strings.Join(submsgs[2:], " ")))
		} else {
			user.SendMsgToUser("用户 " + submsgs[1] + " 不存在")
		}
	} else if submsgs[0] == "私聊" && len(submsgs) == 2 {
		user.SendMsgToUser("私聊信息不能为空")
	} else if submsgs[0] == "广播" && len(submsgs) >= 2 {
		user.server.BroadCast(user, msg)
	} else {
		user.SendMsgToUser("服务器错误，消息模式无法识别")
	}
}

func NewUser(conn net.Conn, server *Server) (user *User) {
	UserAddr := conn.RemoteAddr().String()
	user = &User{
		Name:   UserAddr,
		Add:    UserAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
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
