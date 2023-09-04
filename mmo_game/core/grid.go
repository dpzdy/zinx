package core

import (
	"fmt"
	"sync"
)

/*
一个AOI地图中的格子类型
*/
type Grid struct {
	//格子ID
	GID int
	//格子的左边边界坐标
	MinX int
	//格子的右边边界坐标
	MaxX int
	//格子的上边边界坐标
	MinY int
	//格子的下边边界坐标
	MaxY int
	//当前各自内成员的ID集合
	playerIDs map[int]bool
	//保护当前集合的锁
	pIDLock sync.RWMutex
}

// 初始化当前格子的方法
func NewGrid(gID, minX, maxX, minY, maxY int) *Grid {
	return &Grid{
		GID:       gID,
		MinX:      minX,
		MaxX:      maxX,
		MinY:      minY,
		MaxY:      maxY,
		playerIDs: make(map[int]bool),
	}
}

// 给格子添加一个玩家
func (g *Grid) Add(playerID int) {
	g.pIDLock.Lock()
	defer g.pIDLock.Unlock()

	g.playerIDs[playerID] = true
}

// 从格子删除一个玩家
func (g *Grid) Remove(playerID int) {
	g.pIDLock.Lock()
	defer g.pIDLock.Unlock()

	delete(g.playerIDs, playerID)
}

// 得到当前格子中所有的玩家
func (g *Grid) GetPlayerIDs() (playIDs []int) {
	g.pIDLock.RLock()
	defer g.pIDLock.RUnlock()
	for k, _ := range g.playerIDs {
		playIDs = append(playIDs, k)
	}
	return
}

// 调试-打出格子的基本信息
func (g *Grid) String() string {

	return fmt.Sprintf("Grid id: %3d,minX: %3d,maxX: %3d,minY: %3d,maxY: %3d,playerIDs: %v",
		g.GID, g.MinX, g.MaxX, g.MinY, g.MaxY, g.playerIDs)
}
