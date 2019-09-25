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
	Item     *Item  `json:"item,omitempty"`
}

type Request struct {
	Type string `json:"type"`
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

func (node *Node) StoreItem() {
	inMemoryItems[node.Item.ID] = *node.Item
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
	/* accept connection on port */
	conn, err := ln.Accept()
	resp := Response{}
	if err != nil {
		resp.Status = "KO"
		resp.Message = "Something went wrong, impossible to accept connection, port" + node.Port
	} else {
		var req Request
		json.NewDecoder(conn).Decode(&req)
		fmt.Printf("Request: %v", req)
		fmt.Printf("Item: %v", node.Item)

		switch req.Type {
		case "GET":
			node.CountItems()
			resp.Status = "OK"
		case "POST":
			node.StoreItem()
			resp.Status = "OK"
		default:
			resp.Status = "KO"
			resp.Message = "Sorry, only POST method is supported"
		}

		json.NewEncoder(conn).Encode(&resp)
		conn.Close()
	}
	return err
}

func (node *Node) Clean() {
	node.Item = nil
}
