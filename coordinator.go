package main

import (
	"distributed-counter/nodehandler"
	"strconv"
	"net"
	"encoding/json"
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

	//run nodes:
	for i := 0; i < n-1; i++ {
		go nodes[i].Run()
	}
	nodes[n-1].Run()

}


func (cdt *Coordinator) SendMessages() {
	for _, port := range cdt.Ports {
		conn, err := net.Dial("tcp", ":" + port)
		if err != nil {
			//handle error
		} else {
			var req = nodehandler.Request{ "GET" }
			json.NewEncoder(conn).Encode(&req)
		}
	}
}

func main() {

}