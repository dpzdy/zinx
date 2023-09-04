package core

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"sync"
	"zinx/mmo_game/pb"
	"zinx/zinx/ziface"
)

// 玩家对象
type Player struct {
	Pid  int32              //玩家ID
	Conn ziface.IConnection //当前玩家的链接（用于和客户端的链接）
	X    float32            //平面的X坐标
	Y    float32            //高度
	Z    float32            //平面Y坐标（注意）
	V    float32            //旋转的0-360角度
}

/*
Player ID 生成器
*/
var PidGen int32 = 1
var IdLock sync.Mutex

func NewPlayer(conn ziface.IConnection) *Player {
	//生成一个玩家ID
	IdLock.Lock()
	id := PidGen
	PidGen++
	IdLock.Unlock()

	//创建一个玩家对象
	p := &Player{
		Pid:  id,
		Conn: conn,
		X:    float32(160 + rand.Intn(10)),
		Y:    0,
		Z:    float32(140 + rand.Intn(20)),
		V:    0,
	}
	return p
}

/*
提供一个发送个客户端消息的方法
主要是将pbde protobuf数据序列化后，在调用sendmsg方法
*/
func (p *Player) SendMsg(msgId uint32, data proto.Message) {
	//将protco Message 结构体序列化 转换成二进制
	msg, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("marshal msg err ", err)
		return
	}

	//将二进制文件 通过zinx框架的sendmsg将数据发送给客户端
	if p.Conn == nil {
		fmt.Println("connection in player is nil")
		return
	}
	if err := p.Conn.SendMsg(msgId, msg); err != nil {
		fmt.Println("Player sendmsg err!")
		return
	}
	return
}

// 将playerID同步给客户端	给客户端发送msgid：1 的消息
func (p *Player) SyncPid() {
	//组建MsgID：1的protoc数据
	proto_msg := &pb.SyncPid{
		Pid: p.Pid,
	}
	p.SendMsg(1, proto_msg)
}

// 广播玩家自己的出生地点 将player的上线的初始位置同步给客户端	给客户端发送msgid：200 的消息
func (p *Player) BroadCastStartPosition() {
	//组建MsgID：200的protoc数据
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //代表广播位置坐标
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}
	p.SendMsg(200, proto_msg)
}

// 玩家广播世界聊天消息
func (p *Player) Talk(content string) {
	//组建msgid200 proto数据
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  1,
		Data: &pb.BroadCast_Content{
			content,
		},
	}
	//得到当前世界所有的在线玩家
	players := WorldMgrObj.GetAllPlayers()
	//向所有的玩家发送msgid200的消息
	for _, player := range players {
		//player分别给对应的客户端发消息
		player.SendMsg(200, proto_msg)
	}
}

// 同步玩家上线的位置消息
func (p *Player) SyncSurrounding() {
	//获取当前玩家周围的玩家有哪些
	pids := WorldMgrObj.AoiMgr.GetPidsByPos(p.X, p.Z)
	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}
	//将当前的位置信息通过msgid:202发个周围的玩家(让其他玩家看到自己)
	//	2.1 组建msgid200 proto 数据
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}
	//	2.2 全部周围的玩家都向格子的客户端发送200消息,proto_msg
	for _, player := range players {
		player.SendMsg(200, proto_msg)
	}
	//将周围的全部玩家的位置信息发送给当前的玩家客户端(让自己看到其他玩家)msgid:202
	//3.1 制作msgid 20 proto数据
	//3.1.1 制作pb.Player slice
	players_proto_msg := make([]*pb.Player, 0, len(players))
	for _, player := range players {
		//制作一个message Player
		p := &pb.Player{
			Pid: player.Pid,
			P: &pb.Position{
				X: player.X,
				Y: player.Y,
				Z: player.Z,
				V: player.V,
			},
		}
		players_proto_msg = append(players_proto_msg, p)
	}
	//3.1.2 封装suncplayer protobuf数据
	SuncPlayers_proto_msg := &pb.SyncPlayers{
		Ps: players_proto_msg[:],
	}
	//3.2 将组建好的数据发给当前客户端
	p.SendMsg(202, SuncPlayers_proto_msg)

}

// 广播当前玩家的位置移动信息
func (p *Player) UpdatePos(x, y, z, v float32) {
	//更新当前玩家player对象的坐标
	p.X, p.Y, p.Z, p.V = x, y, z, v
	//组件广播proto协议 msgid:200 tp-4
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  4,
		Data: &pb.BroadCast_P{
			&pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}
	//获取当前玩家的周边玩家AOI九宫格之内的玩家
	players := p.GetSurroundPlayers()
	//给每个玩家对应的客户端发送当前玩家位置更新的信息
	for _, player := range players {
		player.SendMsg(200, proto_msg)
	}
}
func (p *Player) GetSurroundPlayers() []*Player {
	pids := WorldMgrObj.AoiMgr.GetPidsByPos(p.X, p.Z)
	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}
	return players
}

// 玩家下线
func (p *Player) OffLine() {
	//得到当前玩家周边的九宫格内的玩家
	players := p.GetSurroundPlayers()
	//给周围玩家广播msgid201
	proto_msg := &pb.SyncPid{
		Pid: p.Pid,
	}
	for _, player := range players {
		player.SendMsg(201, proto_msg)
	}
	//将当前玩家从世界管理器删除
	WorldMgrObj.RemovePlayerByPid(p.Pid)
	//将当前玩家从AOI管理器删除
	WorldMgrObj.AoiMgr.RemoveFromGridByPos(int(p.Pid), p.X, p.Z)
}
