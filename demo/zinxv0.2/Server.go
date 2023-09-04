package main

import "zinx/zinx/znet"

/*
基于zinx框架开发的  服务器段应用程序
*/

func main() {
	//1.创建一个server句柄，使用zinx的api
	s := znet.NewServer("[zinx V0.2]")
	//2.启动server
	s.Serve()
}
