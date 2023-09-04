package main

import (
	"fmt"
	"zinx/zinx/ziface"
	"zinx/zinx/znet"
)

/*
基于zinx框架开发的  服务器段应用程序
*/

// ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

// Test PreHandle
func (this *PingRouter) PreHandle(request ziface.IRequest) {
	fmt.Println("Call Router PreHadnle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("before ping...\n"))
	if err != nil {
		fmt.Println("call back before ping error")
	}
}

// Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router Hadnle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("ping... ping... ping...\n"))
	if err != nil {
		fmt.Println("call back before ping error")
	}
}

// Test PostHandle
func (this *PingRouter) PostHandle(request ziface.IRequest) {
	fmt.Println("Call Router PostHadnle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("After ping...\n"))
	if err != nil {
		fmt.Println("call back after ping error")
	}
}

func main() {
	//1.创建一个server句柄，使用zinx的api
	s := znet.NewServer("[zinx V0.3]")
	//2.给当前zinx添加一个自定义的router
	s.AddRouter(&PingRouter{})
	//3.启动server
	s.Serve()
}
