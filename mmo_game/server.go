package main

import (
	"fmt"
	"zinx/mmo_game/apis"
	"zinx/mmo_game/core"
	"zinx/zinx/ziface"
	"zinx/zinx/znet"
)

// 当前客户端建立链接之后的hook函数
func OnConnectionAdd(conn ziface.IConnection) {
	//创建一个player对象
	player := core.NewPlayer(conn)
	//给客户端发送msgid：1 的消息  同步Player的ID给客户端
	player.SyncPid()
	//给客户端发送msgid：200 的消息 同步Player的位置给客户端
	player.BroadCastStartPosition()
	//将当前新上线的玩家添加到WordManager中
	core.WorldMgrObj.AddPlayer(player)
	//将该链接绑定一个pid 玩家ID属性
	conn.SetProperty("pid", player.Pid)

	//同步周边玩家,广播自己已经上线
	player.SyncSurrounding()

	fmt.Println("==========>Player pid = ", player.Pid, " is arrived <===============")
}

// 当前客户端断开链接之前的hook函数
func OnConnectionLost(conn ziface.IConnection) {
	//通过链接属性得到当前链接绑定的pid
	pid, _ := conn.GetProperty("pid")
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))
	//触发玩家下线业务
	player.OffLine()
	fmt.Println("==========>Player pid = ", pid, " offline <===============")

}
func main() {
	//创建zinx server句柄
	s := znet.NewServer("MMO Game Zinx")
	//链接创建销毁的hook函数
	s.SetOnConnStart(OnConnectionAdd)
	s.SetOnConnStop(OnConnectionLost)
	//注册一些路由业务
	s.AddRouter(2, &apis.WorldChatApi{})
	s.AddRouter(3, &apis.MoveApi{})
	//启动服务
	s.Serve()
}
