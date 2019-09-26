package main

import (
	"distributed-counter/nodehandler"
	"fmt"
	"testing"
	"time"
)

func TestNodesInitialization(t *testing.T) {
	cdt := Coordinator{}
	go cdt.StartNodeInstances(2)
	time.Sleep(2 * time.Second)
}

func TestItemsBalancing(t *testing.T) {
	balancedIdx := GetBalancedIndexes(13, 5)
	fmt.Printf("%v\n", balancedIdx)
}

func TestNodesCommunication(t *testing.T) {
	item := nodehandler.Item{ID: 1, Tenant: "teeest"}
	ports := []string{"8090"}
	node := nodehandler.Node{0, ports[0], true, false}
	cdt := Coordinator{[]*nodehandler.Node{&node}}
	go node.Run()
	time.Sleep(2 * time.Second)
	cdt.WriteMessagesToNodes(nodehandler.Request{Type: "POST"}, []*nodehandler.Item{&item})
	time.Sleep(2 * time.Second)
}

func TestWithMoreItemsThanNodes(t *testing.T) {
	cdt := Coordinator{}
	go cdt.StartNodeInstances(2)
	items := []*nodehandler.Item{
		&nodehandler.Item{ID: 1, Tenant: "hello"},
		&nodehandler.Item{ID: 2, Tenant: "world"},
		&nodehandler.Item{ID: 3, Tenant: "hello"},
		&nodehandler.Item{ID: 4, Tenant: "public"},
		&nodehandler.Item{ID: 5, Tenant: "sonar"},
	}
	time.Sleep(2 * time.Second)
	go cdt.WriteMessagesToNodes(nodehandler.Request{Type: "POST"}, items)
	time.Sleep(5 * time.Second)
}

func TestGetCounter(t *testing.T) {
	node1 := nodehandler.Node{NodeID: 1, Port: "8090", IsMaster: true}
	node2 := nodehandler.Node{NodeID: 1, Port: "8091", IsMaster: false}
	cdt := Coordinator{Nodes: []*nodehandler.Node{&node1, &node2}}

	go node1.Run()
	go node2.Run()
	time.Sleep(1 * time.Second)

	items := []*nodehandler.Item{
		&nodehandler.Item{ID: 1, Tenant: "PublicSonar"},
		&nodehandler.Item{ID: 2, Tenant: "PublicSonar"},
	}

	cdt.CallNode(nodehandler.Request{Type: "POST"}, &node2, items)

	counter, err := cdt.GetCounter("PublicSonar")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("Counter: %v\n", *counter)
}
