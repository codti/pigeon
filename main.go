package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"pigeon/lib/config"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func main() {

	config.Init()
	host := viper.GetString("base.host")
	port := viper.GetString("base.port")
	network := viper.GetString("base.network")
	service := fmt.Sprintf("%s:%s", host, port)

	// log.Println(service)
	// 设置日志
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	// 解析地址
	addr, _ := net.ResolveTCPAddr(network, service)

	// 创建套接字
	listener, err := net.ListenTCP(network, addr)

	if err != nil {
		log.Println("create lesten err", err)
		os.Exit(0)
	}

	// 程序退出关闭套接字
	defer listener.Close()

	// 等待接收连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener.Accept err", err)
		}

		go connectHandle(conn)
	}
}

// connectHandle 连接处理函数
func connectHandle(conn net.Conn) {
	defer conn.Close()
	// 定义src来源存储数据容器
	buff := make([]byte, 1024)

	n, err := conn.Read(buff)

	// 来源数据读取成功
	if err == nil {
		header := strings.Split(string(buff), "\r\n")
		isConnect := strings.HasPrefix(header[0], "CONNECT")

		// 处理目的相关操作
		host, port, err := parseHeader(string(buff))
		log.Println(host)
		if err != nil {
			log.Println(host, port, err)
		}

		dstConn, err := net.Dial("tcp", host+":"+port)
		if err != nil {
			log.Println(err)
			return
		}

		dstConn.SetDeadline(time.Now().Add(6 * time.Second))

		// 建立CONNECT通道
		if isConnect {
			conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		} else {
			n, err = dstConn.Write(buff[:n])
			if err != nil {
				log.Println(err)
			}
		}

		ExitChan := make(chan bool, 1)
		go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
			_, err := io.Copy(dconn, sconn)
			if err != nil {
				fmt.Printf("往%v发送数据失败:%v end\n", host, err)
			}
			ExitChan <- true
		}(conn, dstConn, ExitChan)
		go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
			_, err := io.Copy(sconn, dconn)
			if err != nil {
				fmt.Printf("从%v接收数据失败:%v\n", host, err)
			}
			ExitChan <- true
		}(conn, dstConn, ExitChan)

		<-ExitChan

		defer dstConn.Close()
	} else {
		log.Println(err)
	}
}

// 获取host 和 port
func parseHeader(str string) (string, string, error) {
	head := strings.Split(str, "\r\n")
	host, port, err := net.SplitHostPort(strings.Replace(head[1], "Host:", "", -1))

	if err != nil {
		log.Println("head=", head[1], err)
		// http 初始化
		addr := strings.Split(strings.Replace(head[1], "Host:", "", -1), ":")
		host = addr[0]
		if len(addr) == 1 {
			port = "80"
		} else {
			port = addr[1]
		}
	}
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if len(port) == 0 {
		port = "80"
	}
	log.Println("host=", host, ";port=", port)
	return host, port, nil
}
