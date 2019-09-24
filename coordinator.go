package main

import (
	"distributed-counter/nodehandler"
	"strconv"
)

type Coordinator struct {
	Ports		[]string							`json:"ports`
	Nodes		[]*nodehandler.Node		`json:"nodes"`
}

var cdt = Coordinator{}

func (cdt *Coordinator) StartNodeInstances(n int, items []nodehandler.Item) {
	nodes := make([]*nodehandler.Node, n)

	ports := make([]string, n)
	initialPort := 8090

	for i := 0; i < n; i++ {
		ports[i] = strconv.Itoa(initialPort)
		initialPort++
	}

	//Initialize nodes:
	for i := 0; i < n; i++ {
		isMaster := i == 0
		nodes[i] = &nodehandler.Node{i, ports[i], isMaster, items[i]}
	}

	for i := 0; i < n; i++ {
		go nodes[i].Run()
	}
}


func (cdt *Coordinator) SendMessages() {
	
}

func main() {

}