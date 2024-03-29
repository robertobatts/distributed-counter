package main

import (
	"distributed-counter/nodehandler"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"errors"

	"github.com/gorilla/mux"
)

type Coordinator struct {
	Nodes []*nodehandler.Node `json:"nodes"`
}

var cdt = Coordinator{}
var NODES_NUMBER = 4

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

func GetBalancedIndexes(itemsDim int, inodesDim int) []int {
	itemsPerNode := itemsDim / inodesDim
	reminder := itemsDim % inodesDim
	balancedIdxs := make([]int, inodesDim+1)
	balancedIdxs[0] = 0
	for i := 1; i < len(balancedIdxs); i++ {
		balancedIdxs[i] = balancedIdxs[i-1] + itemsPerNode
		//distribute the reminder equally across nodes
		if reminder > 0 {
			balancedIdxs[i]++
			reminder--
		}
	}
	return balancedIdxs
}

func (cdt *Coordinator) WriteMessagesToNodes(req nodehandler.Request, items []*nodehandler.Item) {

	done := make(chan bool, len(cdt.Nodes)-1)
	balancedIdxs := GetBalancedIndexes(len(items), len(cdt.Nodes))

	for i := 0; i < len(cdt.Nodes); i++ {
		go func(idx int) {
			if !cdt.Nodes[idx].IsBusy {
				cdt.CallNode(req, cdt.Nodes[idx], items[balancedIdxs[idx]:balancedIdxs[idx+1]])
			} else {
				//listening continuosly for available nodes until I don't find one
				cdt.CallAvailableNode(req, items[balancedIdxs[idx]:balancedIdxs[idx+1]])
			}
			done <- true
		}(i)
	}

	for i := 0; i < len(done); i++ {
		<-done
	}
}

func (cdt *Coordinator) CallNode(req nodehandler.Request, node *nodehandler.Node, items []*nodehandler.Item) (*nodehandler.Response, error) {
	node.IsBusy = true
	/*get the master before calling it for GET operations, so that I can compare it then
	to see if the master node successfully connected to another port after crashing
	*/
	master := cdt.GetMasterNode()
	conn, err := net.Dial("tcp", ":"+node.Port)

	if err != nil {
		fmt.Println(err.Error())
		switch(req.Type) {
		case "POST":		
			cdt.CallAvailableNode(req, items)
		case "GET":	
			/*if the master crashes, it changes port. I check if the port changed so that I can connect
			to it again and the system is still query-able*/
			if master.Port != node.Port {
				cdt.CallNode(req, node, items)
			}
		}
	} else {
		req.MasterPort = master.Port
		if req.Type == "POST" {
			req.Items = items
		}

		errEnc := json.NewEncoder(conn).Encode(&req)
		var resp nodehandler.Response
		errDec := json.NewDecoder(conn).Decode(&resp)
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

func (cdt *Coordinator) CallAvailableNode(req nodehandler.Request, items []*nodehandler.Item) {
	for {
		newNode := cdt.GetAvailableNode()
		if newNode != nil {
			_, err := cdt.CallNode(req, newNode, items)
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

func (cdt *Coordinator) GetCounter(tenant string) (*int, error) {
	node := cdt.GetMasterNode()
	req := nodehandler.Request{Type: "GET", Tenant: tenant}
	resp, err := cdt.CallNode(req, node, nil)
	if err != nil {
		return nil, err
	}

	return resp.Counter, nil
}

func ValidateItems(items []*nodehandler.Item) error {
	for _, item := range items {
		if item.ID == nil || item.Tenant == nil || *item.Tenant == "" {
			return errors.New("Items not valid, ID or tenant missing")
		}
	}
	return nil
}

func Items(w http.ResponseWriter, req *http.Request) {
	resp := nodehandler.Response{}
	switch req.Method {
	case "GET":
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
		items := []*nodehandler.Item{}
		err := decoder.Decode(&items)

		if err != nil {
			resp.Status = "KO"
			resp.Message = err.Error()
		} else {
			err := ValidateItems(items)
			if err != nil {
				resp.Status = "KO"
				resp.Message = err.Error()
			} else {
				cdt.WriteMessagesToNodes(nodehandler.Request{Type: "POST"}, items)
				resp.Status = "OK"
			}
		}
	default:
		resp.Status = "KO"
		resp.Message = "Sorry, only POST and GET methods are supported"
	}
	if (resp.Status != "OK") {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
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
