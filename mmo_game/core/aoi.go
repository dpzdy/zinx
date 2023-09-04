package core

import (
	"fmt"
)

// 定义一些AOI的边界值
const (
	AOI_MIN_X  int = 85
	AOI_MAX_X  int = 410
	AOI_CNTS_X int = 10
	AOI_MIN_Y  int = 75
	AOI_MAX_Y  int = 400
	AOI_CNTS_Y int = 20
)

/*
AOI管理模块
*/

type AOIManager struct {
	//区域的左边界坐标
	MinX int
	//区域的右边界坐标
	MaxX int
	//X方向格子的数量
	CntsX int
	//区域的上边界坐标
	MinY int
	//区域的下边界坐标
	MaxY int
	//Y方向格子的数量
	CntsY int
	//当前区域中有哪些格子map-key=格子的ID,value=格子对象
	grids map[int]*Grid
}

/*
初始化一个AOI区域管理模块
*/
func NewAOIManager(minX, maxX, cntsX, minY, maxY, cntsY int) *AOIManager {
	aoiMgr := &AOIManager{
		MinX:  minX,
		MaxX:  maxX,
		MinY:  minY,
		MaxY:  maxY,
		CntsX: cntsX,
		CntsY: cntsY,
		grids: make(map[int]*Grid),
	}
	//给AOI初始化区域的格子所有的格子进行标号和初始化
	for y := 0; y < cntsY; y++ {
		for x := 0; x < cntsX; x++ {
			//计算格子的ID 根据xy编号  id = idy*cntX + idx
			gid := y*cntsX + x
			//初始化gid格子
			aoiMgr.grids[gid] = NewGrid(gid,
				aoiMgr.MinX+x*aoiMgr.gridWidth(),
				aoiMgr.MinX+(x+1)*aoiMgr.gridWidth(),
				aoiMgr.MinY+y*aoiMgr.gridLength(),
				aoiMgr.MinY+(y+1)*aoiMgr.gridLength())
		}
	}

	return aoiMgr
}

// 得到每个格子在X轴方向的宽度
func (m *AOIManager) gridWidth() int {
	return (m.MaxX - m.MinX) / m.CntsX

}

// 得到每个格子在Y轴方向的长度
func (m *AOIManager) gridLength() int {
	return (m.MaxY - m.MinY) / m.CntsY
}

// 打印信息
func (m *AOIManager) String() string {
	//打印AOIManager信息
	s := fmt.Sprintf("AOIManager:\nMinX:%d, MaxX:%d, CntsX:%d, MinY:%d, MaxY:%d, CntsY:%d\nAOIGrids:\n",
		m.MinX, m.MaxX, m.CntsX, m.MinY, m.MaxY, m.CntsY)
	//打印全部格子信息
	for _, grid := range m.grids {
		s += fmt.Sprintln(grid)
	}
	return s
}

// 根据GID得到周围九宫格格子集合
func (m *AOIManager) GetSurroundGridsByGid(gID int) (grids []*Grid) {
	//判断当前gID是否在AOIManager中
	if _, ok := m.grids[gID]; !ok {
		return
	}
	//初始化grids返回值切片
	grids = append(grids, m.grids[gID])

	//判断gid左右是否有格子 gid -> x
	idx := gID % m.CntsX
	//左边是否有格子，放在gidsx集合中
	if idx > 0 {
		grids = append(grids, m.grids[gID-1])
	}
	//右边是否有格子，放在gidsx集合中
	if idx < m.CntsX-1 {
		grids = append(grids, m.grids[gID+1])
	}
	//将x轴当前格子取出
	gridsx := make([]int, 0, len(grids))
	for _, v := range grids {
		gridsx = append(gridsx, v.GID)
	}
	//判断gid上下是否有格子 gid -> y
	for _, v := range gridsx {
		idy := v / m.CntsY
		//上边是否有格子，放在gids集合中
		if idy > 0 {
			grids = append(grids, m.grids[v-m.CntsX])
		}
		//下边是否有格子，放在gids集合中
		if idy < m.CntsY-1 {
			grids = append(grids, m.grids[v+m.CntsX])
		}
	}
	return
}

// 得到当前玩家的GID格子Id
func (m *AOIManager) GetGidByPos(x, y float32) int {
	idx := (int(x) - m.MinX) / m.gridWidth()
	idy := (int(y) - m.MinY) / m.gridLength()

	return idy*m.CntsX + idx

}

// 通过坐标等到周围九宫格全部的playersIDs
func (m *AOIManager) GetPidsByPos(x, y float32) (playersIDs []int) {
	//得到当前玩家的GID格子Id
	gID := m.GetGidByPos(x, y)
	//通过gid得到周边九宫格的信息
	grids := m.GetSurroundGridsByGid(gID)
	//将九宫格信息里的全部player的id添加到playerid
	for _, v := range grids {
		playersIDs = append(playersIDs, v.GetPlayerIDs()...)
		//fmt.Printf("====> grid ID : %d, pids : %v\n", v.GID, v.GetPlayerIDs())
	}
	return
}

// 添加一个playerID到一个格子中
func (m *AOIManager) AddPidToGrid(pID, gID int) {
	m.grids[gID].Add(pID)
}

// 移除一个格子中的playerID
func (m *AOIManager) RemovePidFromGrid(pID, gID int) {
	m.grids[gID].Remove(pID)
}

// 通过GID获取全部的playerID
func (m *AOIManager) GetPidsByGid(gID int) (playerIDs []int) {
	playerIDs = m.grids[gID].GetPlayerIDs()
	return
}

// 通过坐标将player添加到一个格子中
func (m *AOIManager) AddToGridByPos(pID int, x, y float32) {
	gID := m.GetGidByPos(x, y)
	grid := m.grids[gID]
	grid.Add(pID)
}

// 通过坐标将player从一个格子中删除
func (m *AOIManager) RemoveFromGridByPos(pID int, x, y float32) {
	gID := m.GetGidByPos(x, y)
	grid := m.grids[gID]
	grid.Remove(pID)
}

//protoc  62
