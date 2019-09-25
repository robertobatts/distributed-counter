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
	Nodes []*nodehandler.Node `json:"nodes"`
}

var cdt = Coordinator{}
var NODES_NUMBER = 2

func (cdt *Coordinator) StartNodeInstances(n int) {
	nodes := make([]*nodehandler.Node, n)

	port := 8090

	//Initialize nodes:
	for i := 0; i < n; i++ {
		isMaster := i == 0
		nodes[i] = &nodehandler.Node{i, strconv.Itoa(port), isMaster, false}
		port++
	}
	cdt.Nodes = nodes

	//nodes start listening for the coordinator:
	for i := 0; i < n-1; i++ {
		go nodes[i].Run()
	}
	nodes[n-1].Run()

}

func (cdt *Coordinator) WriteMessagesToNodes(req nodehandler.Request, items []nodehandler.Item) {
	done := make(chan bool, len(items)-1)

	for i := 0; i < len(items); i++ {
		go func(idx int) {
			if idx < len(cdt.Nodes) {
				cdt.CallNode(req, cdt.Nodes[idx], &items[idx])
			} else {
				//listening continuosly for available nodes until I don't find one
				cdt.CallAvailableNode(req, items[idx])
			}
			done <- true
		}(i)
	}

	for i := 0; i < len(done); i++ {
		<-done
	}
}

func (cdt *Coordinator) CallNode(req nodehandler.Request, node *nodehandler.Node, item *nodehandler.Item) (*nodehandler.Response, error) {
	node.IsBusy = true
	conn, err := net.Dial("tcp", ":"+node.Port)
	if err != nil {
		fmt.Println(err.Error())
		if req.Type == "POST" {
			cdt.CallAvailableNode(req, *item)
		}
	} else {
		req.Item = item

		errEnc := json.NewEncoder(conn).Encode(&req)
		var resp nodehandler.Response
		errDec := json.NewDecoder(conn).Decode(&resp)
		fmt.Printf("Status: %v\n", resp.Status)
		conn.Close()
		if errEnc == nil && errDec == nil && resp.Status == "OK" {
			/*if the data has been written correctly, I cancel the item from the Node,
			so that the coordinator can recognize that the node is available to work another item*/
			node.IsBusy = false
			return &resp, nil
		} else {
			if errEnc != nil {
				return nil, errEnc
			}
			if errDec != nil {
				return nil, errDec
			}
		}
	}
	return nil, nil
}

func (cdt *Coordinator) CallAvailableNode(req nodehandler.Request, item nodehandler.Item) {
	for {
		newNode := cdt.GetAvailableNode()
		if newNode != nil {
			_, err := cdt.CallNode(req, newNode, &item)
			if err == nil {
				break
			}
		}
	}
}

func (cdt *Coordinator) GetAvailableNode() *nodehandler.Node {
	for _, node := range cdt.Nodes {
		if !node.IsBusy {
			return node
		}
	}
	return nil
}

func (cdt *Coordinator) GetMasterNode() *nodehandler.Node {
	for _, node := range cdt.Nodes {
		if node.IsMaster {
			return node
		}
	}
	return nil
}

func (cdt *Coordinator) AlignNodesMemory() {
	//no need to align if there is only one node
	if len(cdt.Nodes) > 1 {
		master := cdt.GetMasterNode()
		req := nodehandler.Request{Type: "MOVE", MasterPort: master.Port}
		//no reasono to run it parallel, because the master port will be busy
		for _, node := range cdt.Nodes {
			if !node.IsMaster {
				cdt.CallNode(req, node, nil) //TODO: handle the error
			}
		}
	}
}

func (cdt *Coordinator) GetCounter(tenant string) (int, error) {
	node := cdt.GetMasterNode()
	req := nodehandler.Request{Type: "GET"}
	resp, err := cdt.CallNode(req, node, &nodehandler.Item{Tenant: tenant})
	if err != nil {
		return 0, err
	}
	return resp.Counter, nil

}

func Items(w http.ResponseWriter, req *http.Request) {
	resp := nodehandler.Response{}
	switch req.Method {
	case "GET":
		cdt.AlignNodesMemory()
		tenant := mux.Vars(req)["tenant"]
		counter, err := cdt.GetCounter(tenant)

		if err != nil {
			resp.Status = "KO"
			resp.Message = err.Error()
		} else {
			resp.Counter = counter
		}
	case "POST":
		decoder := json.NewDecoder(req.Body)
		items := []nodehandler.Item{}
		err := decoder.Decode(&items)

		if err != nil {
			resp.Status = "KO"
			resp.Message = err.Error()
		} else {
			//balanceNodes() distribute items across nodes
			cdt.WriteMessagesToNodes(nodehandler.Request{Type: "POST"}, items)
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
	go cdt.StartNodeInstances(NODES_NUMBER)

	r := mux.NewRouter()
	r.HandleFunc("/items/{tenant}/count", Items)
	r.HandleFunc("/items", Items)
	http.ListenAndServe(":8085", r)

}
