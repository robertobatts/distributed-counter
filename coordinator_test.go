package main

import (
	"distributed-counter/nodehandler"
	"testing"
	"time"
)

func TestNodesCommunication(t *testing.T) {
	item := nodehandler.Item{1, "teeest"}
	ports := []string{"8090"}
	node := nodehandler.Node{0, ports[0], true, nil}
	cdt := Coordinator{ports, []*nodehandler.Node{&node}}
	go node.Run()
	time.Sleep(2 * time.Second)
	cdt.SendMessagesToNodes(nodehandler.Request{"POST"}, []nodehandler.Item{item})
	time.Sleep(2 * time.Second)
}

func TestNodesInitialization(t *testing.T) {
	cdt := Coordinator{}
	go cdt.StartNodeInstances(2)
	time.Sleep(2 * time.Second)
}
