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
	"time"
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
	
	go func() {
		time.Sleep(time.Millisecond * 300)
		client.Close()
	}()
	
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
	
	go func() {
		time.Sleep(time.Millisecond * 300)
		client.Close()
	}()
	
	err = client.Call("KademliaRPC.FindNode", findNodeRequest, &findNodeResult)
	if err != nil {
		return nil, &CommandFailed{
			"Unable to find node " + fmt.Sprintf("%s:%v", contact.Host.String(), contact.Port)}
	} else {
		k.KB.CommandChannel <- &KBucketRequest{"main", UPDATE, nil, contact, nil}
		for i:=0;i<len(findNodeResult.Nodes);i++{
			k.KB.CommandChannel <- &KBucketRequest{"main", UPDATE, nil, &findNodeResult.Nodes[i],nil}
		}
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

    go func() {
		time.Sleep(time.Millisecond * 300)
		client.Close()
	}()
    
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

func (k *Kademlia) FindNodeRoutine(ct Contact, id ID, chnn chan FindNodeResult) {
	c, err := k.DoFindNode(&ct, id)
	chnn <- FindNodeResult{ct.NodeID, c, err}
}
// For project 2!
func (k *Kademlia) DoIterativeFindNode(id ID) ([]*Contact, error) {
	cs, err := k.DoFindNode(&k.SelfContact, id)
	if err == nil{
		var cs_a []*Contact
		for _, c := range cs {
			cs_a = append(cs_a, &c)
			//fmt.Println("initial ",c.NodeID.AsString())
		}
		sl := new(ShortList)
		sl.initializeShortList(cs_a, id)
		
		chnn := make(chan FindNodeResult)
		var count int
		nocloser := false
		for !sl.checkActive() && !nocloser {
			//fmt.Println("\n\n")
		    sl.printStatus()
		    //fmt.Println("\n\n")
			f3 := sl.getAlphaNotContacted()
			for _, c := range f3 {
				go k.FindNodeRoutine(*c, id, chnn)
				//fmt.Println("alpha ",(*c).NodeID.AsString())
			}
			count = 0
			nocloser = true
			//fmt.Println("---------------------------")
			for count < len(f3) {
				select {
					case fnr := <- chnn:
					     if !sl.checkActive() {
					     	if fnr.Err != nil {
					     		sl.removeInactive(&Contact{fnr.MsgID, nil, 0})
					     	} else {
					     		for i := 0; i < len(fnr.Nodes); i++ {
				//	     			fmt.Println(fnr.Nodes[i].NodeID.AsString())
					     			if sl.updateActiveContact(&fnr.Nodes[i]) {
					     				nocloser = false
					     			}        
					     		}
					     		sl.setActive(&Contact{fnr.MsgID, nil, 0})
					     	}
					     }
					     count++
				    default:
				}
			}
		}
		
		//fmt.Println("\n\n")
		sl.printStatus()
		
		for !sl.checkActive() {
			cs_a = sl.getAllNotContacted()
			if len(cs_a) == 0 {
				break
			}
			count = 0
			for _, c := range cs_a {
				go k.FindNodeRoutine(*c, id, chnn)
		//		fmt.Println("NotContacted ",(*c).NodeID.AsString())
			}
			
			for count < len(cs_a) {
				select {
					case fnr := <- chnn:
					     if !sl.checkActive() {
					     	if fnr.Err == nil {
					     		for i := 0; i < len(fnr.Nodes); i++ {
					     			sl.updateActiveContact(&fnr.Nodes[i])        
					     		}
					     		sl.setActive(&Contact{fnr.MsgID, nil, 0})
		//			     		fmt.Println(fnr.MsgID.AsString())
					     	}
					     }
					     count++
				    default:
				}
			}
		}
		return sl.getActiveNodes(), nil
		//return a string slice that contains the ID of the k closest nodes
	} else {
		return nil, &CommandFailed{"Timeout"} 
	}
}

func (k *Kademlia) DoIterativeStore(key ID, value []byte) ([]*Contact, error) {
	cs, err := k.DoIterativeFindNode(key)
	if err != nil {
		return nil, &CommandFailed{"Unable to store key-value pairs iteratively"}
	} else {
		for _, c := range cs {
			k.DoStore(c, key, value)                       
		}
		return cs, nil
	}
}

func (k *Kademlia) FindValueRoutine(ct Contact, key ID, chnn chan FindValueResult) {
	v, c, err := k.DoFindValue(&ct, key)	
	chnn <- FindValueResult{ct.NodeID, v, c, err}
}

func (k *Kademlia) DoIterativeFindValue(key ID) (id string, value []byte, err error) {
	vs, cs, err := k.DoFindValue(&k.SelfContact, key)	//find value at local
	if err == nil {		
		if vs != nil {
			return k.SelfContact.NodeID.AsString(), vs, nil                              //the value is found at local
		} else {
			var cs_a []*Contact
			for _, c := range cs {
				cs_a = append(cs_a, &c)
			}
			sl := new(ShortList)
			sl.initializeShortList(cs_a, key)           //initialize shortlist
			
			chnn := make(chan FindValueResult)
			var count int
			nocloser := false
			var v_found []byte
			var id_found string
			for !sl.checkActive() && !nocloser {
				f3 := sl.getAlphaNotContacted()
				for _, c := range f3 {
					go k.FindValueRoutine(*c, key, chnn)  //find value at the first alpha nodes
				}
				
				count = 0
				nocloser = true
				for count < len(f3) {                                                     ///loop until all routines return results 
					select {
						case fvr := <- chnn:
						     if !sl.checkActive() && v_found == nil {
						     	if fvr.Err != nil {
						     		sl.removeInactive(&Contact{fvr.MsgID, nil, 0})
						     	} else if fvr.Value != nil {
						     		v_found = fvr.Value
						     		id_found = fvr.MsgID.AsString()
						     	} else {
						     		for i := 0; i < len(fvr.Nodes); i++ {
						     			if sl.updateActiveContact(&fvr.Nodes[i]) {     
						     				nocloser = false    				     	
						     			} 
						     		}
						     		sl.setActive(&Contact{fvr.MsgID, nil, 0})   
						     	}						     	
						     }
						     count++
					    default:                                       
					}
				}
				
				if v_found != nil {    
					ID_found,_ := IDFromString(id_found)             
					active := sl.getActiveNodes()
					dis := -1
					var toStore *Contact
					for _,con := range active{
						if !ID_found.Equals(con.NodeID){
							if temp := distance(ID_found,con.NodeID); temp >dis{
								dis = temp
								toStore = con
							}
						}
					}
					k.DoStore(toStore,key,v_found)
					return id_found, v_found, nil                         ///need modification: store the value in the closest node!!!!!!
				}
			}
			
			for !sl.checkActive() {
				cs_a = sl.getAllNotContacted()
				if len(cs_a) == 0 {
					break
				}
				count = 0
				for _, c := range cs_a {
					go k.FindValueRoutine(*c, key, chnn)
				}
				
				for count < len(cs_a) {
					select {
						case fvr := <- chnn:
						     if !sl.checkActive() && v_found == nil {
						     	if fvr.Err != nil {
						     		sl.removeInactive(&Contact{fvr.MsgID, nil, 0})
						     	} else if fvr.Value != nil {
						     		v_found = fvr.Value
						     		id_found = fvr.MsgID.AsString()
						     	} else {
						     		for i := 0; i < len(fvr.Nodes); i++ {
						     			if sl.updateActiveContact(&fvr.Nodes[i]) {     
						     				nocloser = false    				     	
						     			} 
						     		}
						     		sl.setActive(&Contact{fvr.MsgID, nil, 0})   
						     	}						     	
						     }
						     count++				         
				        default:
					}
				}
				if v_found != nil {
					ID_found,_ := IDFromString(id_found)             
					active := sl.getActiveNodes()
					dis := -1
					var toStore *Contact
					for _,con := range active{
						if !ID_found.Equals(con.NodeID){
							if temp := distance(ID_found,con.NodeID); temp >dis{
								dis = temp
								toStore = con
							}
						}
					}
					k.DoStore(toStore,key,v_found)
					return id_found, v_found, nil                                ///need modification: store the value in the closest node!!!!
				}
			}
			clst := sl.getClosestNodes()

			return clst[0], nil, &CommandFailed{"Key not found"} 
		}
	} else {
		return k.SelfContact.NodeID.AsString(), nil, &CommandFailed{"Timeout"}
	}	
}

// For project 3!
func (k *Kademlia) Vanish(data []byte, numberKeys byte,
	threshold byte, timeoutSeconds int) (vdo VanashingDataObject) {
	return
}

func (k *Kademlia) Unvanish(searchKey ID) (data []byte) {
	return nil
}
