package main

import (
	"distributed-counter/nodehandler"
	"testing"
	"time"
)

func TestNodesInitialization(t *testing.T) {
	cdt := Coordinator{}
	go cdt.StartNodeInstances(2)
	time.Sleep(2 * time.Second)
}

func TestNodesCommunication(t *testing.T) {
	item := nodehandler.Item{1, "teeest"}
	ports := []string{"8090"}
	node := nodehandler.Node{0, ports[0], true, false, nil}
	cdt := Coordinator{[]*nodehandler.Node{&node}}
	go node.Run()
	time.Sleep(2 * time.Second)
	cdt.SendMessagesToNodes(nodehandler.Request{"POST"}, []nodehandler.Item{item})
	time.Sleep(2 * time.Second)
}

func TestWithMoreItemsThanNodes(t *testing.T) {
	cdt := Coordinator{}
	go cdt.StartNodeInstances(2)
	items := []nodehandler.Item{
		nodehandler.Item{1, "hello"},
		nodehandler.Item{2, "world"},
		nodehandler.Item{3, "hello"},
		nodehandler.Item{4, "public"},
		nodehandler.Item{5, "sonar"},
	}
	time.Sleep(2 * time.Second)
	go cdt.SendMessagesToNodes(nodehandler.Request{"POST"}, items)
	time.Sleep(5 * time.Second)
}
