/*****************************************************************************/
/**
* \file       fwd.go
* \author     gaoxj
* \date       2018/05/31
* \brief      本地任务控制
* \note       Copyright (c) 2000-2020  赛特斯信息科技股份有限公司
* \remarks    修改日志
******************************************************************************/

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type SocketServer struct {
	onosSocketMap    map[string]*net.TCPConn
	netconfSocketMap map[string]*net.TCPConn
	sn               string
}

var socketServer *SocketServer
var completeOnos chan int = make(chan int)
var completeNetconf chan int = make(chan int)

//var devSn string = "1"

func socketConfInit(config *ConfigData) {

	for i := 1; i < 4; i++ {
		if i%8 == 0 && i >= 8 {
			time.Sleep(10 * time.Second)
		}
		devSn := strconv.Itoa(i)
		messageExchange(devSn, config)
	}
	<-completeOnos
	<-completeNetconf
}

func (sok *SocketServer) dealMessage(sn string, socket *net.TCPConn, con chan (int), config *ConfigData) {
	var remoteSocket *net.TCPConn
	//var dataSource string
	if s, ok := sok.onosSocketMap[sn]; ok {
		if socket == s {
			remoteSocket = sok.netconfSocketMap[sn]
			//dataSource = " onos:"
		} else {
			remoteSocket = sok.onosSocketMap[sn]
			//dataSource = " netconf:"
		}
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error: %s", r)
			if s, ok := sok.onosSocketMap[sn]; ok {
				delete(sok.onosSocketMap, sn)
				delete(sok.netconfSocketMap, sn)
				if socket == s || remoteSocket == s {
					socket.Close()
					remoteSocket.Close()
					time.Sleep(10 * time.Second)
					messageExchange(sn, config)
				}
			}
		}
	}()

	buf := make([]byte, 1024) //定义一个切片的长度是1024
	for true {
		data, err := socket.Read(buf)
		if err != nil {
			panic("a bug occur")
		}
		remoteSocket.Write(buf[:data])
		//fmt.Println("receive data from"+dataSource, string(buf[:data]))

	}
	con <- 0
}

func NewSocketServer(onosSocket *net.TCPConn, netconfSocket *net.TCPConn, sn string) *SocketServer {
	sok := &SocketServer{}
	sok.onosSocketMap = make(map[string]*net.TCPConn)
	sok.netconfSocketMap = make(map[string]*net.TCPConn)
	sok.onosSocketMap[sn] = onosSocket
	sok.netconfSocketMap[sn] = netconfSocket
	sok.sn = sn
	return sok
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: %s", err.Error())
		os.Exit(1)
	}
}
func messageExchange(sn string, config *ConfigData) {
	onosaddr, err := net.ResolveTCPAddr("tcp4", config.onosAddr)
	checkError(err)
	onosSocket, err1 := net.DialTCP("tcp", nil, onosaddr) //作为客户端连接onos，需要指定协议
	checkError(err1)

	netconfaddr, err2 := net.ResolveTCPAddr("tcp4", config.netconfAddr)
	checkError(err2)
	netconfSocket, err3 := net.DialTCP("tcp", nil, netconfaddr) //作为客户端连接netconf，需要指定协议
	checkError(err3)
	socketServer = NewSocketServer(onosSocket, netconfSocket, sn)
	onosSocket.Write([]byte("ENG:" + sn + "\r\n"))
	fmt.Print(sn)
	go socketServer.dealMessage(socketServer.sn, onosSocket, completeOnos, config)
	go socketServer.dealMessage(socketServer.sn, netconfSocket, completeNetconf, config)
}
