package main

import (
	"testing"
	"distributed-counter/nodehandler"
)

func TestNodesCommunication (t *testing.T) {

}

func TestNodesInitialization (t *testing.T) {
	cdt := Coordinator{}
	items := []nodehandler.Item{ nodehandler.Item{1, "teeest"} }
	cdt.StartNodeInstances(1, items)
}