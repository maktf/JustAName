package libkademlia

type VDObj struct {
	kademlia                               *Kademlia
	table                                  map[ID]VanashingDataObject
	DoStoreVDORequest                      chan *VanashingDataObject
	DoGetVDORequestWithGetVDOResultChan    chan *GetVDORequestWithGetVDOResultChan
}

// type Kademlia struct {
// 	NodeID      ID
// 	SelfContact Contact
// 	RM          *RequestManager
// 	KB          *KBuckets
// 	VDO         *VDObj
// }

// type GetVDORequest struct {
// 	Sender Contact
// 	VdoID  ID
// 	MsgID  ID
// }

// type GetVDOResult struct {
// 	MsgID ID
// 	VDO   VanashingDataObject
// }

// type GetVDORequestWithGetVDOResultChan struct {
// 	getVDORequest        GetVDORequest
// 	getVDOResultChan     chan GetVDOResult
// }

func (kademlia *Kademlia) StoreVDOGetVDOHandle () error {
	go func() {
		for {
			select {
				case doStoreVDORequest := <- kademlia.VDO.DoStoreVDORequest: {
					kademlia.StoreVDO(doStoreVDORequest)
					return nil
				}
				case doGetVDORequestWithGetVDOResultChan := <- kademlia.VDO.DoGetVDORequestWithGetVDOResultChan: {
					contact = doGetVDORequestWithGetVDOResultChan.getVDORequest.Sender
					client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+strconv.Itoa(int(contact.Port)),rpc.DefaultRPCPath+strconv.Itoa(int(contact.Port)))
					if err != nil {
						return &CommandFailed{"Uable to dail", err}
					}
					var reply error
					err = client.Call("KademliaRPC.GetVDO", doGetVDORequestWithGetVDOResultChan, &reply)
					if err != nil {
						return &CommandFailed{"Uable to KademliaRPC.GetVDO", err}
					}
					if reply != nil {
						return &CommandFailed{err}
					}
				}
			}
		}
	} ()
}
		// type VanashingDataObject struct {
		// 	AccessKey  int64
		// 	Ciphertext []byte
		// 	NumberKeys byte
		// 	Threshold  byte
		// }
func (kademlia *Kademlia) StoreVDO(value VanashingDataObject) error {
	key := value.AccessKey
	kademlia.VDO.table[key] = value
	checkValue, ok := kademlia.VDO.table[key]
	if ok && checkValue == value {
		return nil
	} else {
		return &CommandFailed{"Uable to store StoreVDO pair"}
	}
}