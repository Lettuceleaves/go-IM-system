package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

type Client struct {
	ServerIp  string
	ServerPot int
	Name      string
	conn      net.Conn
	mode      string
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址, 默认是127.0.0.1")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口, 默认是8888")
	flag.Parse()
}

func New() (client *Client) {
	client = &Client{
		ServerIp:  serverIp,
		ServerPot: serverPort,
		Name:      serverIp,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return
}

func CommandList() {
	fmt.Println("1. 改名，请输入“改名 你的新名字”")
	fmt.Println("2. 查看用户列表，请输入“用户列表”")
	fmt.Println("3. 退出，请输入“退出”")
	fmt.Println("4. 切换模式，请输入“切换 模式”, 模式有“私聊”和“广播”")
	fmt.Println("4. 重新阅读此菜单，请输入“菜单”")
}

func (client *Client) ClientI(exit chan bool) {
	menu := true
	reader := bufio.NewReader(os.Stdin)
	for {
		if menu {
			CommandList()
			menu = false
		}
		lineBytes, isPrefix, err := reader.ReadLine()
		input := string(lineBytes)
		if err == nil && isPrefix {
			fmt.Println("输入过长，请重新输入")
			continue
		}
		if input == "退出" {
			fmt.Println("退出客户端")
			exit <- true
			return
		} else if input == "菜单" {
			menu = true
			continue
		} else if len([]rune(input)) >= 2 && string([]rune(input)[:2]) == "切换" {
			mode := string([]rune(input)[3:])
			if mode == "私聊" {
				client.mode = "私聊"
			} else if mode == "广播" {
				client.mode = "广播"
			} else {
				fmt.Println("输入错误，请重新输入")
			}
			continue
		} else {
			_, err := client.conn.Write([]byte(client.mode + " " + input))
			if err != nil {
				fmt.Println("发送消息失败， 请重试")
			}
		}
	}
}

func (client *Client) ClientO(exit chan bool) {
	buffer := make([]byte, 1024)
	for {
		n, err := client.conn.Read(buffer)
		if err != nil {
			fmt.Println("conn.Read error:", err)
			exit <- true
			return
		}
		fmt.Println(string(buffer[:n-1]))
	}
}

func (client *Client) menu() {
	fmt.Println("1. 私聊模式，请输入“私聊”")
	fmt.Println("2. 广播模式，请输入“广播”")
	fmt.Println("3. 退出，请输入“退出”")
	fmt.Println("请输入模式：")
	reader := bufio.NewReader(os.Stdin)
	lineBytes, isPrefix, err := reader.ReadLine()
	if err != nil || isPrefix {
		fmt.Println("读取程序错误 ")
		return
	}
	client.mode = string(lineBytes)
	for len(client.mode) == 2 {
		fmt.Println("请输入正确格式")
		lineBytes, isPrefix, err = reader.ReadLine()
		if err != nil || isPrefix {
			fmt.Println("读取程序错误 ")
			return
		}
	}
}

func (client *Client) Run() {
	exit := make(chan bool)
	client.menu()
	go client.ClientO(exit)
	go client.ClientI(exit)
	select {
	case <-exit:
		return
	}
}

func main() {
	client := New()
	if client == nil {
		fmt.Println("连接服务器失败")
		return
	}
	fmt.Println("连接服务器成功")
	exit := make(chan bool)
	client.Run()
	select {
	case <-exit:
		return
	}
}
