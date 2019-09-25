package nodehandler

import (
	"encoding/json"
	"fmt"
	"net"
)

/* Informations about node */
type Node struct {
	NodeID   int    `json:"nodeId"`
	Port     string `json:"port"`
	IsMaster bool   `json:"isMaster"`
	IsBusy   bool   `json:"isBusy"`
	Item     *Item  `json:"item,omitempty"`
}

type Request struct {
	Type string `json:"type"`
	Item *Item  `json:"item,omitempty"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Item struct {
	ID     int    `json:"id"`
	Tenant string `json:"tenant"`
}

var inMemoryItems = make(map[int]Item)

func (node *Node) StoreItem(item Item) {
	inMemoryItems[node.Item.ID] = item
}

func (node *Node) CountItems() {
	if node.IsMaster {

	}
}

func (node *Node) Run() error {

	fmt.Println(node)

	err := node.ListenOnPort()

	return err
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
			case "GET":
				node.CountItems()
				resp.Status = "OK"
			case "POST":
				fmt.Printf("Item: %v\n", *req.Item)
				node.StoreItem(*req.Item)
				resp.Status = "OK"
			default:
				resp.Status = "KO"
				resp.Message = "Sorry, only POST method is supported"
			}

			json.NewEncoder(conn).Encode(&resp)
			conn.Close()
		}
	}
	return nil
}

func (node *Node) Clean() {
	node.Item = nil
}
