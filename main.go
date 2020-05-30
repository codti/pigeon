package main

import (
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {

	// 创建套接字
	listener, err := net.Listen("tcp", "0.0.0.0:9501")

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
	buff := make([]byte, 10240)

	n, err := conn.Read(buff)
	// 来源数据读取成功
	if err == nil {
		// 处理目的相关操作
		host, port, err := parseHeader(string(buff))
		if err != nil {
			log.Println(host, port, err)
		}
		log.Println(host)
		dstConn, err := net.Dial("tcp", host+":"+port)
		if err != nil {
			log.Println(err)
			return
		}
		dstConn.SetDeadline(time.Now().Add(2 * time.Second))
		n, err = dstConn.Write(buff[:n])
		if err != nil {
			log.Println(err)
		}
		throwData(dstConn, conn)
	} else {
		log.Println(err)
	}
}

// throwData 数据搬运
func throwData(srcConn, dstConn net.Conn) {
	// defer srcConn.Close()
	defer dstConn.Close()

	buff := make([]byte, 1024000)

	for {
		n, err := srcConn.Read(buff)
		if err != nil || n == 0 {
			log.Println(err)
			break
		}
		n, err = dstConn.Write(buff[:n])
		if err != nil || n == 0 {
			log.Println(err)
			break
		}
	}
}

func parseHeader(str string) (string, string, error) {
	head := strings.Split(str, "\r\n")
	hostArr := strings.Split(head[1], ":")
	host := DeletePreAndSufSpace(hostArr[1])
	port := "80"
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
