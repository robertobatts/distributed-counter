package nodehandler

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

/* Informations about node */
type Node struct {
	NodeID   int    `json:"nodeId"`
	Port     string `json:"port"`
	IsMaster bool   `json:"isMaster"`
	IsBusy   bool   `json:"isBusy"`
}

type Request struct {
	Type       string         `json:"type"`
	Item       *Item          `json:"item,omitempty"`
	MasterPort string         `json:"masterPort"`
	Memory     map[int64]Item `json:"memory"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Item struct {
	ID           int64     `json:"id"`
	Tenant       string    `json:"tenant"`
	LastUpdateDt time.Time `JSON:"lastUpdateDt"`
}

var inMemoryItems = make(map[int64]Item)

func (node *Node) StoreItem(item Item) {
	item.LastUpdateDt = time.Now()
	inMemoryItems[item.ID] = item
}

func MoveDataToMaster(port string) {
	conn, err := net.Dial("tcp", ":"+port)
	if err != nil {
		//handle
	}
	json.NewEncoder(conn).Encode(&Request{Type: "MOVE", Memory: inMemoryItems})
	var resp Response
	json.NewDecoder(conn).Decode(&resp)
	fmt.Printf("Status: %v\n", resp.Status)
	conn.Close()
}

func GetDataFromSlave(req Request) {
	fmt.Println(req.Memory)
}

func (node *Node) HandleMoveRequest(req Request, resp *Response) {
	if !node.IsMaster {
		MoveDataToMaster(req.MasterPort)
	} else {
		GetDataFromSlave(req)
	}
}

func (node *Node) ListenOnPort() error {
	/* Listen for incoming messages */
	ln, _ := net.Listen("tcp", ":"+node.Port)
	//TODO: handle listener error
	for {
		/* accept connection on port */
		conn, err := ln.Accept()
		resp := Response{}
		if err != nil {
			resp.Status = "KO"
			resp.Message = "Something went wrong, impossible to accept connection, port" + node.Port
			return err
		} else {
			var req Request
			json.NewDecoder(conn).Decode(&req)

			switch req.Type {
			case "MOVE":
				node.HandleMoveRequest(req, &resp)

			case "GET":
				//count items
			case "POST":
				fmt.Printf("Item: %v\n", *req.Item)
				node.StoreItem(*req.Item)
				resp.Status = "OK"
			default:
				resp.Status = "KO"
				resp.Message = "Sorry, only POST and GET methods are supported"
			}

			json.NewEncoder(conn).Encode(&resp)
			conn.Close()
		}
	}
	return nil
}

func (node *Node) Run() error {

	fmt.Println(node)

	err := node.ListenOnPort()

	return err
}
