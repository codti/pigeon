package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {

	network := "tcp"
	service := "0.0.0.0:9501"
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

		if err != nil {
			log.Println(host, port, err)
		}

		dstConn, err := net.Dial("tcp", host+":"+port)
		if err != nil {
			log.Println(err)
			return
		}

		dstConn.SetDeadline(time.Now().Add(2 * time.Second))

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

func parseHeader(str string) (string, string, error) {
	head := strings.Split(str, "\r\n")
	host, port, err := net.SplitHostPort(strings.Replace(head[1], "Host:", "", -1))

	if err != nil {
		log.Println("head =", head[1])
	}
	host = DeletePreAndSufSpace(host)
	port = DeletePreAndSufSpace(port)
	if len(port) == 0 {
		port = "80"
	}
	return host, port, nil
}

// DeletePreAndSufSpace 去除空格
func DeletePreAndSufSpace(str string) string {
	strList := []byte(str)
	spaceCount, count := 0, len(strList)
	for i := 0; i <= len(strList)-1; i++ {
		if strList[i] == 32 {
			spaceCount++
		} else {
			break
		}
	}

	strList = strList[spaceCount:]
	spaceCount, count = 0, len(strList)
	for i := count - 1; i >= 0; i-- {
		if strList[i] == 32 {
			spaceCount++
		} else {
			break
		}
	}

	return string(strList[:count-spaceCount])
}
