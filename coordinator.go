package main

import (
	"distributed-counter/nodehandler"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Coordinator struct {
	Ports []string            `json:"ports`
	Nodes []*nodehandler.Node `json:"nodes"`
}

var cdt = Coordinator{}

func (cdt *Coordinator) StartNodeInstances(n int) {
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
		nodes[i] = &nodehandler.Node{i, ports[i], isMaster, nil}
	}

	//nodes start listening for the coordinator:
	for i := 0; i < n-1; i++ {
		go nodes[i].Run()
	}
	nodes[n-1].Run()

}

func (cdt *Coordinator) SendMessagesToNodes(req nodehandler.Request, items []nodehandler.Item) {
	for i, item := range items {
		go func() {
			if i < len(cdt.Ports) {
				cdt.Nodes[i].Item = &item
				conn, err := net.Dial("tcp", ":"+cdt.Ports[i])
				if err != nil {
					//handle error
				} else {
					errEnc := json.NewEncoder(conn).Encode(&req)
					var resp nodehandler.Response
					errDec := json.NewDecoder(conn).Decode(&resp)
					fmt.Printf("Status: %v", resp.Status)
					if errEnc != nil && errDec != nil && resp.Status == "OK" {
						/*if the data is been written correctly, I cancel the item from the Node,
						so that the coordinator can recognize that the node is available to work another item*/
						cdt.Nodes[i].Item = nil
					} else {
						//TODO: try with other ports
					}
				}
			} else {
				//getAvailableNode(), listening continuosly for available nodes
			}
		}()
	}
}

func (cdt *Coordinator) GetAvailableNode() *nodehandler.Node {
	for _, node := range cdt.Nodes {
		if node.Item == nil {
			return node
		}
	}
	return nil
}

func Items(w http.ResponseWriter, req *http.Request) {
	resp := nodehandler.Response{}
	switch req.Method {
	case "GET":
		//tenant := mux.Vars(req)["tenant"]
		//TODO: handle GET
	case "POST":
		decoder := json.NewDecoder(req.Body)
		items := []nodehandler.Item{}
		err := decoder.Decode(&items)

		if err != nil {
			resp.Status = "KO"
			resp.Message = err.Error()
		} else {
			//balanceNodes() distribute items across nodes
			cdt.SendMessagesToNodes(nodehandler.Request{"POST"}, items)
			resp.Status = "OK"
		}
	default:
		resp.Status = "KO"
		resp.Message = "Sorry, only POST and GET methods are supported"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	go cdt.StartNodeInstances(6)

	r := mux.NewRouter()
	r.HandleFunc("/items/{tenant}/count", Items)
	r.HandleFunc("/items", Items)
	http.ListenAndServe(":8085", r)

}
