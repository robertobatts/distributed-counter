package main

import (
	"testing"
	"distributed-counter/nodehandler"
)

func TestNodesCommunication (t *testing.T) {
	item := nodehandler.Item{ 1, "teeest" }
	ports := []string{ "8090",}
	node := nodehandler.Node{0, ports[0], true, item}
	cdt := Coordinator{ ports, []*nodehandler.Node{ &node } }
	cdt.SendMessages()
}

func TestNodesInitialization (t *testing.T) {
	cdt := Coordinator{}
	items := []nodehandler.Item{ nodehandler.Item{1, "teeest"}, nodehandler.Item{2, "ciaoo"} }
	cdt.StartNodeInstances(2, items)
}