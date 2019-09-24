package nodehandler

import (
	"testing"
)

func TestNodeHandler(t *testing.T) {
	node := Node{NodeID: 1, Port: "8099"}
	node.Run()
}
