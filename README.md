# distributed-counter

## Run 

```
go build coordinator.go

./coordinator
```

This is how you have to send messages to save items into the nodes:

```bash
POST http://localhost:8085/items
[
    {
        "ID": 1,
        "tenant": "hello"
    },
    {
        "ID": 2,
        "tenant": "publicsonar"
    }
]
```

This is how you retrieve the counter of the number of items grouped by tenant:

```bash
GET http://localhost:8085/items/{tenant}/count
```



## How it works

### Start

When the program starts, the Coordinator initializes 4 nodes in a safe way (if a port is busy, it tries automatically with another). All the nodes listen on a TCP port and one of them is the Master

### Update/Insert (POST /items)

The Coordinator distributes the items among the nodes, then it call the nodes parallelly so that they can save the items in their memory.  After saving the items, the nodes become again available to be called by the Coordinator while they're sending their local memory to the Master. This is an optimization, in fact the user doesn't have to wait for the data replication to receive his http response, although the data are replicated at every request.

#### Load Balancing

The items are distributed equally among the nodes. For example, if I have 4 nodes and 11 items the Coordinator distributes them in this way: 

```
Node1: 3,    Node2: 3,        Node3: 3,       Node4: 2
```



#### Error handling

- If a node is already busy when the Coordinator calls it, it start looking for an available node to call, until it found one
- If a slave node crashes, there is no data loss because the items are saved in the Master after every request
- If the items sent by the user are not valid, the Coordinator sends an http error

 #### Storage handling

The items are collected into a map (key=ID, value=item). At the moment of the storage the *LastUpdateDt* is saved in the item as an additional field, so that if multiple nodes send the same ID simultaneously with different tenants, the Master node can understand which is the last updated value to save in its memory.

### Read (GET /items/{tenant}/count)

The Coordinator calls the Master node to received the number of items grouped by tenant. The Master is the only node to be query-able because the items are not replicated in the slave nodes during the POST operations

#### Error handling

- If the master node crashes because of network failure, it starts listening automatically to another port (see func Run()). When the Coordinator tries to call it and receives an error, it detects that the port is been changed and it tries calling the new port. In this way the system is still query-able if the Master crashed during a GET operation 



## What to improve

- Data replication: the data should be replicated to every node. In the current system the data are safe from network failures in the master node, but if it crashes for unexpected reasons the system is no more query-able. This could be fixed by implementing something similar to the functions that I used to merge the Master memory from the slaves. It's just a little more complex because we have to deal with merging all the nodes 