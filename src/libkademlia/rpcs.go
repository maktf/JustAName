package libkademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	"net"
	//"time"
	//"fmt"
)

type KademliaRPC struct {
	kademlia *Kademlia
}

// Host identification.
type Contact struct {
	NodeID ID
	Host   net.IP
	Port   uint16
}

///////////////////////////////////////////////////////////////////////////////
// PING
///////////////////////////////////////////////////////////////////////////////
type PingMessage struct {
	Sender Contact
	MsgID  ID
}

type PongMessage struct {
	MsgID  ID
	Sender Contact
}

func (k *KademliaRPC) Ping(ping PingMessage, pong *PongMessage) error {
	// TODO: Finish implementation
	tid := <- k.kademlia.RM.IdChannel
	pong.MsgID = CopyID(ping.MsgID)	
	// Specify the sender
	pong.Sender = k.kademlia.SelfContact
	//fmt.Println("Address:",&pong.Sender,&k.kademlia.SelfContact)
	// Update contact, etc  
	k.kademlia.KB.CommandChannel <- &KBucketRequest{tid, UPDATE, nil, &ping.Sender, nil}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STORE
///////////////////////////////////////////////////////////////////////////////
type StoreRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
	Value  []byte
}

type StoreResult struct {
	MsgID ID
	Err   error
}

func (k *KademliaRPC) Store(req StoreRequest, res *StoreResult) error {
	// TODO: Implement.
	//for{
		//select{
			tid := <- k.kademlia.RM.IdChannel
			res.MsgID = CopyID(req.MsgID)
			k.kademlia.KB.CommandChannel <- &KBucketRequest{tid, STORE, nil, &req.Sender, &req}
			//for{
			//	select{
			result := <- k.kademlia.RM.ResultChannels[tid]
			if result.flag != true {		
			}
			return nil
				//default:
				//	time.Sleep(5*time.Millisecond)
				//}
			//}
		//default:
		//	time.Sleep(5*time.Millisecond)
		//}
	//}
}

///////////////////////////////////////////////////////////////////////////////
// FIND_NODE
///////////////////////////////////////////////////////////////////////////////
type FindNodeRequest struct {
	Sender Contact
	MsgID  ID
	NodeID ID
}

type FindNodeResult struct {
	MsgID ID
	Nodes []Contact
	Err   error
}

func (k *KademliaRPC) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	// TODO: Implement.

	res.MsgID = CopyID(req.MsgID)
	tid := <- k.kademlia.RM.IdChannel
			//fmt.Println("finding node")
	k.kademlia.KB.CommandChannel <- &KBucketRequest{tid, FINDK, &Contact{req.NodeID, nil, 0}, &req.Sender, nil}
			//fmt.Println("Mission delegated")
	result := <- k.kademlia.RM.ResultChannels[tid]
		//fmt.Println("get result")		
	if result.contacts != nil {		
		for i := 0; i < len(result.contacts); i++ {
			res.Nodes = append(res.Nodes, *result.contacts[i])
		}
	}
	return nil
	
}

///////////////////////////////////////////////////////////////////////////////
// FIND_VALUE
///////////////////////////////////////////////////////////////////////////////
type FindValueRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
}

// If Value is nil, it should be ignored, and Nodes means the same as in a
// FindNodeResult.
type FindValueResult struct {
	MsgID ID
	Value []byte
	Nodes []Contact
	Err   error
}

func (k *KademliaRPC) FindValue(req FindValueRequest, res *FindValueResult) error {
	// TODO: Implement.
	res.MsgID = CopyID(req.MsgID)
	tid := <- k.kademlia.RM.IdChannel
	k.kademlia.KB.CommandChannel <- &KBucketRequest{tid, VALUE, &Contact{req.Key, nil, 0}, &req.Sender, nil}

	result := <- k.kademlia.RM.ResultChannels[tid]
	if result.value != nil {
		for i := 0; i < len(result.value); i++ {
			res.Value = append(res.Value, result.value[i])
		}
	} else if result.contacts != nil {
		for i := 0; i < len(result.contacts); i++ {
			res.Nodes = append(res.Nodes, *result.contacts[i])
		}
	}
	return nil
}

// For Project 3

type GetVDORequest struct {
	Sender Contact
	VdoID  ID
	MsgID  ID
}

type GetVDOResult struct {
	MsgID ID
	VDO   VanashingDataObject
}
// type Kademlia struct {
// 	sync.RWMutex
// 	NodeID      ID
// 	SelfContact Contact
// 	RM          *RequestManager
// 	KB          *KBuckets
// 	VDOs        map[ID]VanashingDataObject
// }
func (k *KademliaRPC) GetVDO(req GetVDORequest, res *GetVDOResult) error {
	// TODO: Implement.
	// package main

	// import "fmt"

	// func main() {
	//         dict := map[string]int {"foo" : 1, "bar" : 2}
	//         value, ok := dict["baz"]
	//         if ok {
	//                 fmt.Println("value: ", value)
	//         } else {
	//                 fmt.Println("key not found")
	//         }
	// }
	key := req.VdoID
	value, ok := k.kademlia.VDOs[key]
	if ok {
		res.MsgID = req.MsgID
		res.VDO = value
		return nil
	} else {
		return &CommandFailed{"VdoID not found"}
	}
}
