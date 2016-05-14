package libkademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID      ID
	SelfContact Contact
	RM          *RequestManager
	KB          *KBuckets
}

func NewKademliaWithId(laddr string, nodeID ID) *Kademlia {
	k := new(Kademlia)
	k.NodeID = nodeID
    k.RM = new(RequestManager)
    k.KB = new(KBuckets)
    k.RM.ResultChannels = make(map[string]chan *KResult)
	k.RM.IdChannel = make(chan string)
	send := make(chan *KResult, 10)
	k.RM.ResultChannels["main"] = send
	for i:=0;i<10240;i++{
			send = make(chan *KResult, 10)
		    k.RM.ResultChannels[string(i)] = send
	}
    k.KB.CommandChannel = make(chan *KBucketRequest, 100)
	// TODO: Initialize other state here as you add functionality.
    
	// Set up RPC server
	// NOTE: KademliaRPC is just a wrapper around Kademlia. This type includes
	// the RPC functions.

	s := rpc.NewServer()
	s.Register(&KademliaRPC{k})
	hostname, port, err := net.SplitHostPort(laddr)
	if err != nil {
		return nil
	}
	s.HandleHTTP(rpc.DefaultRPCPath+port,
		rpc.DefaultDebugPath+port)
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	// Run RPC server forever.
	go http.Serve(l, nil)

	// Add self contact
	hostname, port, _ = net.SplitHostPort(l.Addr().String())
	port_int, _ := strconv.Atoi(port)
	ipAddrStrings, err := net.LookupHost(hostname)
	var host net.IP
	for i := 0; i < len(ipAddrStrings); i++ {
		host = net.ParseIP(ipAddrStrings[i])
		if host.To4() != nil {
			break
		}
	}
	k.SelfContact = Contact{k.NodeID, host, uint16(port_int)}	
	go k.RM.ManagerStart()
	go k.KB.Run(k, k.RM)
	return k
}

func NewKademlia(laddr string) *Kademlia {
	return NewKademliaWithId(laddr, NewRandomID())
}

type ContactNotFoundError struct {
	id  ID
	msg string
}

func (e *ContactNotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}

func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
	
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	} else {
		k.KB.CommandChannel <- &KBucketRequest{"main", FIND, &Contact{nodeId, nil, 0}, &k.SelfContact, nil}
		result := <- k.RM.ResultChannels["main"]
		if result.contacts != nil {
			return result.contacts[0], nil
		} else {
			return nil, &ContactNotFoundError{nodeId, "Not found"}
		}
	} 	
}

type CommandFailed struct {
	msg string
}

func (e *CommandFailed) Error() string {
	return fmt.Sprintf("%s", e.msg)
}

func (k *Kademlia) DoPing(host net.IP, port uint16) (*Contact, error) {
	// TODO: Implement
	//fmt.Println(host.String()+":"+strconv.Itoa(int(port)))

	client, err := rpc.DialHTTPPath("tcp", host.String()+":"+strconv.Itoa(int(port)),rpc.DefaultRPCPath+strconv.Itoa(int(port)))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var reply PongMessage
	/*fmt.Println("========================")
	for i:= 0;i<IDBytes;i++{
		fmt.Println(int(k.SelfContact.NodeID[i]))
	}
	fmt.Println("========================")*/
	err = client.Call("KademliaRPC.Ping", PingMessage{k.SelfContact, NewRandomID()}, &reply)
	//fmt.Println(reply.MsgID)
	if err == nil {
		k.KB.CommandChannel <- &KBucketRequest{"main", UPDATE, nil, &reply.Sender, nil}
		return &reply.Sender, nil
	} else {
		return nil, &CommandFailed{
		   "Unable to ping " + fmt.Sprintf("%s:%v", host.String(), port)}
	}	
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) error {
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+strconv.Itoa(int(contact.Port)),rpc.DefaultRPCPath+strconv.Itoa(int(contact.Port)))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var reply StoreResult
	/*fmt.Println("========================")
	for i:= 0;i<IDBytes;i++{
		fmt.Println(int(k.SelfContact.NodeID[i]))
	}
	fmt.Println("========================")*/
	err = client.Call("KademliaRPC.Store", StoreRequest{k.SelfContact, NewRandomID(),key,value}, &reply)
	//fmt.Println(reply.MsgID)
	if err == nil {
		k.KB.CommandChannel <- &KBucketRequest{"main", UPDATE, nil, contact, nil}
		return nil
	} else {
		return &CommandFailed{
		   "Unable to store " + fmt.Sprintf("%s:%v", contact.Host.String(), contact.Port)}
	}	
	
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) ([]Contact, error) {
	// TODO: Implement
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+strconv.Itoa(int(contact.Port)),
		rpc.DefaultRPCPath+strconv.Itoa(int(contact.Port)))
	if err != nil {
		log.Fatal("dialing", err)
	} 
	findNodeRequest := FindNodeRequest{k.SelfContact, NewRandomID(), searchKey}
	var findNodeResult FindNodeResult
	err = client.Call("KademliaRPC.FindNode", findNodeRequest, &findNodeResult)
	if err != nil {
		return nil, &CommandFailed{
			"Unable to find node " + fmt.Sprintf("%s:%v", contact.Host.String(), contact.Port)}
	} else {
		k.KB.CommandChannel <- &KBucketRequest{"main", UPDATE, nil, contact, nil}
		//for i:=0;i<len(findNodeResult.Nodes);i++{
		//	fmt.Println(findNodeResult.Nodes[i].NodeID.AsString(),findNodeResult.Nodes[i].Host, findNodeResult.Nodes[i].Port)
		//}
		//fmt.Println("====================",len(findNodeResult.Nodes))
		return findNodeResult.Nodes, nil
	}	
}

func (k *Kademlia) DoFindValue(contact *Contact,
	searchKey ID) (value []byte, contacts []Contact, err error) {
	// TODO: Implement
	//fmt.Println(host.String()+":"+strconv.Itoa(int(port)))

	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+strconv.Itoa(int(contact.Port)),
		rpc.DefaultRPCPath+strconv.Itoa(int(contact.Port)))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var res FindValueResult

	err = client.Call("KademliaRPC.FindValue", FindValueRequest{k.SelfContact, NewRandomID(), searchKey}, &res)
	//fmt.Println(res.MsgID)
	if err == nil {
		k.KB.CommandChannel <- &KBucketRequest{"main", UPDATE, nil, contact, nil}
		return res.Value, res.Nodes, nil
	} else {
		return nil, nil, &CommandFailed{
			"Unable to find value " + fmt.Sprintf("%s:%v", contact.Host.String(), contact.Port)}
	}
}

func (k *Kademlia) LocalFindValue(searchKey ID) ([]byte, error) {
	// TODO: Implement
	
	//fmt.Println(res.MsgID)
	k.KB.CommandChannel <- &KBucketRequest{"main", VALUE, &Contact{searchKey, nil, 0}, &k.SelfContact, nil}
	res := <- k.RM.ResultChannels["main"]
	if res.value != nil {
		return res.value, nil
	} else {
		return []byte(""), &CommandFailed{"Unable to find value at local"}
	}	
}

// For project 2!
func (k *Kademlia) DoIterativeFindNode(id ID) ([]Contact, error) {
	return nil, &CommandFailed{"Not implemented"}
}
func (k *Kademlia) DoIterativeStore(key ID, value []byte) ([]Contact, error) {
	return nil, &CommandFailed{"Not implemented"}
}
func (k *Kademlia) DoIterativeFindValue(key ID) (value []byte, err error) {
	return nil, &CommandFailed{"Not implemented"}
}

// For project 3!
func (k *Kademlia) Vanish(data []byte, numberKeys byte,
	threshold byte, timeoutSeconds int) (vdo VanashingDataObject) {
	return
}

func (k *Kademlia) Unvanish(searchKey ID) (data []byte) {
	return nil
}
