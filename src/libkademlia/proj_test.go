package libkademlia

import (
	"bytes"
	"net"
	"strconv"
	"testing"
	// "fmt"
	"time"
	// "net/http"
	"net/rpc"
)

func StringToIpPort(laddr string) (ip net.IP, port uint16, err error) {
	hostString, portString, err := net.SplitHostPort(laddr)
	if err != nil {
		return
	}
	ipStr, err := net.LookupHost(hostString)
	if err != nil {
		return
	}
	for i := 0; i < len(ipStr); i++ {
		ip = net.ParseIP(ipStr[i])
		if ip.To4() != nil {
			break
		}
	}
	portInt, err := strconv.Atoi(portString)
	port = uint16(portInt)
	return
}

func TestPing(t *testing.T) {
	instance1 := NewKademlia("localhost:7890")
	instance2 := NewKademlia("localhost:7891")
	host2, port2, _ := StringToIpPort("localhost:7891")
	contact2, err := instance2.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("A node cannot find itself's contact info")
	}
	contact2, err = instance2.FindContact(instance1.NodeID)
	if err == nil {
		t.Error("Instance 2 should not be able to find instance " +
			"1 in its buckets before ping instance 1")
	}
	instance1.DoPing(host2, port2)
	contact2, err = instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	wrong_ID := NewRandomID()
	_, err = instance2.FindContact(wrong_ID)
	if err == nil {
		t.Error("Instance 2 should not be able to find a node with the wrong ID")
	}

	contact1, err := instance2.FindContact(instance1.NodeID)
	if err != nil {
		t.Error("Instance 1's contact not found in Instance 2's contact list")
		return
	}
	if contact1.NodeID != instance1.NodeID {
		t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
	}
	if contact2.NodeID != instance2.NodeID {
		t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
	}
	return
}

func TestStore(t *testing.T) {
	// test Dostore() function and LocalFindValue() function
	instance1 := NewKademlia("localhost:7892")
	instance2 := NewKademlia("localhost:7893")
	host2, port2, _ := StringToIpPort("localhost:7893")
	instance1.DoPing(host2, port2)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	key := NewRandomID()
	value := []byte("Hello World")
	err = instance1.DoStore(contact2, key, value)
	if err != nil {
		t.Error("Can not store this value")
	}
	storedValue, err := instance2.LocalFindValue(key)
	if err != nil {
		t.Error("Stored value not found!")
	}
	if !bytes.Equal(storedValue, value) {
		t.Error("Stored value did not match found value")
	}
	return
}

func TestFindNode(t *testing.T) {
	// tree structure;
	// A->B->tree
	/*
	         C
	      /
	  A-B -- D
	      \
	         E
	*/
	instance1 := NewKademlia("localhost:7894")
	instance2 := NewKademlia("localhost:7895")
	host2, port2, _ := StringToIpPort("localhost:7895")
	instance1.DoPing(host2, port2)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	tree_node := make([]*Kademlia, 30)
	for i := 0; i < 10; i++ {
		address := "localhost:" + strconv.Itoa(7896+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)
	}
	//fmt.Println("First round start")
	key := NewRandomID()
	contacts, err := instance1.DoFindNode(contact2, key)
	//fmt.Println("First find return")
	if err != nil {
		t.Error("Error doing FindNode")
	}

	if contacts == nil || len(contacts) == 0 {
		t.Error("No contacts were found")
	}
	// TODO: Check that the correct contacts were stored
	key = tree_node[7].SelfContact.NodeID
	//t.Error("key:",key)
	contacts, err = instance1.DoFindNode(contact2,key)
	if err != nil{
		t.Error("(2) Error doing FindNode")
		return
	}
	for i:= 0;i<len(contacts);i++{
		if contacts[i].NodeID.Equals(key){
			_, err = instance1.DoPing(contacts[0].Host,contacts[0].Port)
			if err != nil{
				t.Error("Return Wrong Contact Value")
			}else{
				return 
			}
		}
	}
	t.Error("No correct contacts returned")
	//       (and no other contacts)

	return
}

func TestFindValue(t *testing.T) {
	// tree structure;
	// A->B->tree
	/*
	         C
	      /
	  A-B -- D
	      \
	         E
	*/
	instance1 := NewKademlia("localhost:7926")
	instance2 := NewKademlia("localhost:7927")
	host2, port2, _ := StringToIpPort("localhost:7927")
	instance1.DoPing(host2, port2)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}

	tree_node := make([]*Kademlia, 30)
	for i := 0; i < 10; i++ {
		address := "localhost:" + strconv.Itoa(7928+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)
	}

	key := NewRandomID()
	value := []byte("Hello world")
	err = instance2.DoStore(contact2, key, value)
	if err != nil {
		t.Error("Could not store value")
	}

	// Given the right keyID, it should return the value
	foundValue, contacts, err := instance1.DoFindValue(contact2, key)
	if !bytes.Equal(foundValue, value) {
		t.Error("Stored value did not match found value")
	}

	//Given the wrong keyID, it should return k nodes.
	wrongKey := NewRandomID()
	foundValue, contacts, err = instance1.DoFindValue(contact2, wrongKey)
	if contacts == nil || len(contacts) < 10 {
		t.Error("Searching for a wrong ID did not return contacts", len(contacts))
	}

	// TODO: Check that the correct contacts were stored
	for i:=0;i<len(contacts);i++ {
		_, err = instance1.DoPing(contacts[i].Host,contacts[i].Port)
		if err != nil{
			t.Error("Wrong contacts were stored")
		}
	}
	//       (and no other contacts)
	return
}

func TestDoIterativeFindNode (t *testing.T) {
	number := 4
	instances := make([]*Kademlia, number)
	for i := 0; i < number; i++ {
		address := "localhost:" + strconv.Itoa(8100 + i)
		instances[i] = NewKademlia(address)
	}
	for i := 1; i < number; i++ {
		host := instances[i - 1].SelfContact.Host
		port := instances[i - 1].SelfContact.Port
		_, err := instances[i].DoPing(host, port)
		if err != nil {
			t.Error("TestDoIterativeFindNode - DoPing - ", err)
		}
	}
	for i := 0; i < number; i++ {
		for j := 0; j < number; j++ {
			contacts, err := instances[i].DoIterativeFindNode(instances[j].NodeID)
			if err != nil {
				t.Error("TestDoIterativeFindNode - DoIterativeFindNode - ", err)
			} else {
				if len(contacts) == k {
					for x := 0; x < len(contacts); x++ {
						client, err := rpc.DialHTTPPath("tcp", contacts[x].Host.String() + ":" + strconv.Itoa(int(contacts[x].Port)), rpc.DefaultRPCPath + strconv.Itoa(int(contacts[x].Port)))
						if err != nil {
							t.Error("TestDoIterativeFindNode - rpc.DialHTTPPath - ", err)
						} else {
							go func() {
								time.Sleep(time.Millisecond * 300)
								client.Close()
							} ()
							findNodeRequest := FindNodeRequest{instances[i].SelfContact, NewRandomID(), instances[j].NodeID}
							var findNodeResult FindNodeResult
							err = client.Call("KademliaRPC.FindNode", findNodeRequest, &findNodeResult)
							if err != nil {
								t.Error("TestDoIterativeFindNode - DoIterativeFindNode - Exist InActive Node - ", err)
								break
							}
						}
					}
				} else {
					var minDistance int
					minDistance = 1<<32 - 1
					for y := 0; y < len(contacts); y++ {
						currentDistance := distance(instances[i].SelfContact.NodeID, contacts[y].NodeID)
						if currentDistance < minDistance {
							minDistance = currentDistance
						}
					}					
					for y := 0; y < len(contacts); y++ {
						returnedContacts, err := instances[i].DoFindNode(contacts[y], instances[j].NodeID)
						if err != nil {
							t.Error("TestDoIterativeFindNode - returnedContacts - DoFindNode - ", err)
						} else {
							for z := 0; z < len(returnedContacts); z++ {
								currentDistance := distance(instances[i].SelfContact.NodeID, returnedContacts[z].NodeID)
								if currentDistance < minDistance {
									t.Error("TestDoIterativeFindNode - Wrong Results")
									break
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestDoIterativeStore (t *testing.T) {
	number := 4
	instances := make([]*Kademlia, number)
	for i := 0; i < number; i++ {
		address := "localhost:" + strconv.Itoa(8200 + i)
		instances[i] = NewKademlia(address)
	}
	for i := 1; i < number; i++ {
		host := instances[i - 1].SelfContact.Host
		port := instances[i - 1].SelfContact.Port
		_, err := instances[i].DoPing(host, port)
		if err != nil {
			t.Error("TestDoIterativeStore - DoPing - ", err)
			break
		}
	}
	keys := make([]ID, number)
	values := make([]Value, number)
	for i := 0; i < number; i++ {
		key := NewRandomID()
		value := Value("TestDoIterativeStore - " + instances[i].NodeID.AsString() + " - to store - key = " + key.AsString())
		keys[i] = key
		values[i] = value
		instances[i].DoStore(&instances[i].SelfContact, key, value)
		// instances[i].DoStore(&instances[number - 1 - i].SelfContact, key, value)
	}
	// returnedValue, _ := instances[0].LocalFindValue(keys[0])
	// if bytes.Equal(returnedValue, values[0]) {
	// 	t.Error("Yuanhui Yang")
	// }
	// func (k *Kademlia) DoIterativeStore(key ID, value []byte) ([]*Contact, error)
	for i := 0; i < number; i++ {
		key := keys[i]
		value := values[i]
		value = Value("Update - "+ string(value))
		values[i] = value
		// values[number - 1 - i] = value
		_, err := instances[i].DoIterativeStore(key, value)
		if err != nil {
			t.Error("TestDoIterativeStore - DoIterativeStore - ", err)
		}
	}
	for i := 0; i < number; i++ {
		key := keys[i]
		value := values[i]
		returnedValue, err := instances[i].LocalFindValue(key)
		if err != nil {
			t.Error("TestDoIterativeStore - DoIterativeStore - LocalFindValue - ", err)
		} else {
			if !bytes.Equal(returnedValue, value) {
				t.Error("TestDoIterativeStore - DoIterativeStore - LocalFindValue - !bytes.Equal", err)
			}
		}
	}
	// NodeIDAsString, err := instances[i].DoIterativeStore(key, value)
	// NodeIDs := make([]ID, 0)
	// for i := 0; i < number; i++ {
	// 	key := keys[i]
	// 	value := values[i]
	// 	value = Value("Update - " + string(value))
	// 	NodeIDAsString, err := instances[i].DoIterativeStore(key, value)
	// 	if err != nil {
	// 		t.Error("TestDoIterativeStore - DoIterativeStore - ", err)
	// 	} else {
	// 		NodeIDAsBytes, err := IDFromString(NodeIDAsString)
	// 		if err != nil {
	// 			t.Error("TestDoIterativeStore - DoIterativeStore - IDFromString - ", err)
	// 		} else {
	// 			NodeIDs = append(NodeIDs, NodeIDAsBytes
	// 		}
	// 	}
	// }
}

func TestDoIterativeFindValue (t *testing.T) {
	number := 4
	instances := make([]*Kademlia, number)
	for i := 0; i < number; i++ {
		address := "localhost:" + strconv.Itoa(8300 + i)
		instances[i] = NewKademlia(address)
	}
	for i := 1; i < number; i++ {
		host := instances[i - 1].SelfContact.Host
		port := instances[i - 1].SelfContact.Port
		_, err := instances[i].DoPing(host, port)
		if err != nil {
			t.Error("TestDoIterativeFindValue - DoPing - ", err)
		} else {
			
		}
	}
}
