package core

import (
	"fmt"
	"testing"
)

func TestNewAOIManager(t *testing.T) {
	aoiMgr := NewAOIManager(0, 250, 5, 0, 250, 5)
	fmt.Println(aoiMgr)
}

func TestAOIManager_GetSurroundGridsByGid(t *testing.T) {
	aoiMgr := NewAOIManager(0, 250, 5, 0, 250, 5)
	for gid, _ := range aoiMgr.grids {
		grids := aoiMgr.GetSurroundGridsByGid(gid)
		fmt.Println("gid:", gid, "gid len:", len(grids))
		gIDs := make([]int, 0, len(grids))
		for _, grid := range grids {
			gIDs = append(gIDs, grid.GID)
		}
		fmt.Println("Surrounding grid IDs ", gIDs)
	}

}
