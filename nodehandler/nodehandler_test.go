package nodehandler

import (
	"testing"
	"time"
)

func TestNodeHandler(t *testing.T) {
	node := Node{NodeID: 1, Port: "8090"}
	node.Run()
}

func TestMemoryMovement(t *testing.T) {
	node1 := Node{NodeID: 1, Port: "8090", IsMaster: true}
	node2 := Node{NodeID: 1, Port: "8091", IsMaster: false}

	go node1.Run()
	go node2.Run()
	time.Sleep(1 * time.Second)
	item := Item{ID: 1, Tenant: "PublicSonar"}
	node2.StoreItems([]*Item{&item})
	MoveDataToMaster("8090")
	time.Sleep(3 * time.Second)
}
