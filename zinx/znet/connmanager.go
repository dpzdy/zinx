package znet

import (
	"errors"
	"fmt"
	"sync"
	"zinx/zinx/ziface"
)

/*
链接管理模块
*/

type ConnManager struct {
	//管理的链接集合
	connections map[uint32]ziface.IConnection
	//保护链接集合的读写锁
	connLock sync.RWMutex
}

// 创建链接
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

// 添加链接
func (connMgr *ConnManager) Add(conn ziface.IConnection) {
	//保护共享资源，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//将Conn加入到ConnManager中
	connMgr.connections[conn.GetConnID()] = conn
	fmt.Println("connID = ", conn.GetConnID(), " add to ConnManager successfully：conn num = ", connMgr.Len())
}

// 删除链接
func (connMgr *ConnManager) Remove(conn ziface.IConnection) {
	//保护共享资源，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//删除链接信息
	delete(connMgr.connections, conn.GetConnID())
	fmt.Println("connID = ", conn.GetConnID(), " remove from ConnManager successfully：conn num = ", connMgr.Len())
}

// 根据Connid获取链接
func (connMgr *ConnManager) Get(connID uint32) (ziface.IConnection, error) {
	//保护共享资源，加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	}
	return nil, errors.New("connection not FOUND")
}

// 得到当前连接总数
func (connMgr *ConnManager) Len() int {
	return len(connMgr.connections)
}

// 清除并终止所有的链接
func (connMgr *ConnManager) ClearConn() {
	//保护共享资源，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//删除conn并停止conn的工作
	for connID, conn := range connMgr.connections {
		//停止
		conn.Stop()
		//删除
		delete(connMgr.connections, connID)
	}
	fmt.Println("Clear All Connections Succ! conn num = ", connMgr.Len())
}
