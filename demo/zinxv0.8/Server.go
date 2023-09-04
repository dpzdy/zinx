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

// HelloZinx 自定义路由
type HelloZinxRouter struct {
	znet.BaseRouter
}

// Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Hadnle")
	//先读取客户端数据  在写回 ping..ping..ping
	fmt.Println("recv from client : msgId = ", request.GetMsgId(),
		", data = ", string(request.GetData()))
	err := request.GetConnection().SendMsg(200, []byte("ping..ping..ping"))
	if err != nil {
		fmt.Println(err)
	}
}
func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call HelloRouter Hadnle")
	//先读取客户端数据  在写回 ping..ping..ping
	fmt.Println("recv from client : msgId = ", request.GetMsgId(),
		", data = ", string(request.GetData()))
	err := request.GetConnection().SendMsg(201, []byte("Hello zinx"))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	//1.创建一个server句柄，使用zinx的api
	s := znet.NewServer("[zinx V0.8]")
	//2.给当前zinx添加一个自定义的router
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})
	//3.启动server
	s.Serve()
}
