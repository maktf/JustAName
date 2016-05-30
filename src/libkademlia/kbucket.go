package libkademlia

import(
	"fmt"
	"math/rand"
	"time"
)

type Value []byte

type KResult struct{
	contacts []*Contact
	flag bool
	value []byte
}

type RequestManager struct{
	ResultChannels map[string]chan *KResult
	IdChannel chan string
}

type KBucketRequest struct{
	Id string
	Command int
	Contact *Contact
	Sender *Contact
	StoreReq *StoreRequest
}

type KBuckets struct{
	Kademlia *Kademlia
	Buckets [IDBits][]*Contact
	CommandChannel chan *KBucketRequest
	Manager *RequestManager

	LocalTable map[ID]Value
}

//type Bucket []*Contact

const UPDATE = 1
const FIND =2
const FINDK = 3
const STORE = 4
const VALUE = 5

func (manager *RequestManager) ManagerStart(){
	i := 0
	for{
		newid := string(i)
		i++
		i = i%10240
        manager.IdChannel <- newid
	}
}

func (kb *KBuckets) Run(k *Kademlia, man *RequestManager) bool{
	kb.Kademlia = k
	for i:= 0;i<IDBytes;i++ {
		kb.Buckets[i] = make([]*Contact,0,IDBytes)
	}
	kb.LocalTable = make(map[ID]Value)
	kb.Manager = man
	for {
		select{
		case command := <- kb.CommandChannel:
			id := command.Id
			comType := command.Command
			contact := command.Contact
			sender := command.Sender
			//fmt.Println("Command Type: %s",comType)
			switch comType{
			case UPDATE:
				res := new(KResult)
				res.flag = kb.Update(sender, &(k.SelfContact))
				//kb.Manager.ResultChannels[id] <- res

			case FIND:
				res := kb.FindContact(contact.NodeID)
				kb.Update(sender,&(k.SelfContact))
				kb.Manager.ResultChannels[id] <- res

			case FINDK:
				//fmt.Println("start FINDK")
				res := kb.FindKClosest(contact.NodeID,sender.NodeID)
				//fmt.Println("start Update after FINDK")
				kb.Update(sender,&(k.SelfContact))
				//fmt.Println("Fin update")
				kb.Manager.ResultChannels[id] <- res
				//fmt.Println("res put")

			case STORE:
				req := command.StoreReq
				res := kb.Store(req)
				kb.Update(sender,&(k.SelfContact))
				kb.Manager.ResultChannels[id] <- res

			case VALUE:
				res:= kb.FindValue(contact,sender)
				kb.Update(sender,&(k.SelfContact))
				kb.Manager.ResultChannels[id] <- res
			default:
				fmt.Println("Invalid Command")
			}
		default:
			time.Sleep(50*time.Millisecond)
		}
	}
}

func (kb *KBuckets) Store(storeReq *StoreRequest) *KResult{
	res := new(KResult)
	kb.LocalTable[storeReq.Key] = make([]byte,len(storeReq.Value),cap(storeReq.Value))
	copy(kb.LocalTable[storeReq.Key],storeReq.Value)
	res.flag = true
	return res
}

func (kb *KBuckets) FindValue(node *Contact,sender *Contact) *KResult{
	if elem, ok:=kb.LocalTable[node.NodeID]; ok == true{
		res := new(KResult)
		res.value = make([]byte, len(elem),cap(elem))
		copy(res.value,elem)
		res.flag = true
		return res
	}else{
		res := kb.FindKClosest(node.NodeID,sender.NodeID)
		res.flag = false
		return res
	}
}

func (kb *KBuckets) Update(node *Contact, self *Contact) bool{
	dis := distance(node.NodeID, self.NodeID)
	if dis == 160 {
		return true
	}
	//targetBucket := kb.Buckets[dis]
	res := update(node,kb.Kademlia, kb, dis)
	//fmt.Println("Update!! New length of buckets",dis,len(kb.Buckets[dis]))
	return res
} 


func update(node *Contact, k *Kademlia, kb *KBuckets, dis int) bool{
	handler := kb.Buckets[dis]
	if node == nil{
		fmt.Println("NIL error caused by Contact node")
		return false
	}
	if len(handler) == IDBytes{
		for i,v := range handler{
			res,_ := k.DoPing(v.Host, v.Port)
			if res == nil{
				nouvel := make([]*Contact,0,IDBytes)
				for j:= 0;j<i;j++{
					nouvel = append(nouvel,handler[j])
				}
				for j:=i+1;j<len(handler);j++{
					nouvel = append(nouvel,handler[j])
				}
				nouvel = append(nouvel,node)
				//fmt.Println( v.Host, " was removed")
				//fmt.Println(node.NodeID.AsString(), "Updated")
				kb.Buckets[dis] = nouvel
				return true
			}
		}
	}else{
		for i,v := range handler{
			if comp:=compareNodeId(v.NodeID,node.NodeID);comp==0{
				nouvel := make([]*Contact,0,IDBytes)
				for j:= 0;j<i;j++{
					nouvel = append(nouvel,handler[j])
				}
				for j:=i+1;j<len(handler);j++{
					nouvel = append(nouvel,handler[j])
				}
				nouvel = append(nouvel,node)
				//fmt.Println(node.NodeID.AsString(), "Updated")
				kb.Buckets[dis] = nouvel
				return true
			}
		}
		//fmt.Println(node.NodeID.AsString(), "Added")
		handler = append(handler,node)
		kb.Buckets[dis] = handler
	}
	return true
}

func (kb *KBuckets) FindKClosest(key ID,senderId ID) *KResult{
	//fmt.Println("Start finding K closest")
	res:= new(KResult)
	dis := distance(key, kb.Kademlia.SelfContact.NodeID)
	i := IDBytes
	res.contacts = make([]*Contact,0,IDBytes)
	temp := dis-1
	if dis == 160 && !senderId.Equals(kb.Kademlia.SelfContact.NodeID){
		res.contacts = append(res.contacts, &kb.Kademlia.SelfContact)
		i--
	}
	for i>0 && dis <160 {
		for j:=0; i>0 && j<len(kb.Buckets[dis]);j++{
			if !kb.Buckets[dis][j].NodeID.Equals(senderId){
				res.contacts = append(res.contacts, kb.Buckets[dis][j])
				i--
			}
		}
		dis++
	}
	//fmt.Println("Reach the bottom of the kbuckets")
	for i>0 && temp>=0{
		for j:= 0;i>0 && j<len(kb.Buckets[temp]);j++{
			if !kb.Buckets[temp][j].NodeID.Equals(senderId){
				res.contacts = append(res.contacts, kb.Buckets[temp][j])
				i--
			}
		}
		temp--
	}
	//fmt.Println("finish")
	if len(res.contacts) == 0{
		res.flag = false
		res.contacts = nil
	}else{
		res.flag = true
	}
	return res

}


func (kb *KBuckets) FindContact(nodeId ID) *KResult{
	//fmt.Println("===============")
	//fmt.Println("targetNode: ", nodeId)
	res := new(KResult)
	dis := distance(nodeId, kb.Kademlia.NodeID)
	if dis == 160{
		res.contacts = []*Contact{&kb.Kademlia.SelfContact}
		res.flag = true
		return res
	}
	targetBucket := kb.Buckets[dis]
	var comp int
	for _,v := range targetBucket{
		if comp = compareNodeId(v.NodeID,nodeId); comp == 0{
			res.flag = true
			res.contacts = []*Contact{v}
			//fmt.Println("Successfully Found")
			//fmt.Println("=================")
			return res
		}else{
			//fmt.Println(v.NodeID,"Failed")
		}
	}
	res.flag = false
	res.contacts = nil
	//fmt.Println("===================")
	return res
}

func distance(nodeId ID, selfId ID) int {
	var dis [IDBytes]byte
	for i := 0;i<IDBytes;i++{
		dis[i] = nodeId[i]^selfId[i]
		//fmt.Println(int(dis[i]),int(nodeId[i]),int(selfId[i]))
	}
	var i int
	var j uint
	for i = 0 ;i<IDBytes;i++{
		for j = 0;j<8;j++{
			if (128>>j)<=dis[i]{
				//fmt.Println("Calculated Distance: ", i*8+int(j), i, j)
				return i*8+int(j)
			}
		}
	}
	//fmt.Println("Calculated Distance: (self detected) ", i*8+int(j), i, j)
	return 160
}



func compareNodeId(id1 ID, id2 ID) int{
	for i,v := range id1{
		if v > id2[i]{
			return 1
		}else if v<id2[i]{
			return -1
		}
	}
	return 0
}


func NewChannelID() (ret []byte) {
	for i := 0; i < IDBytes; i++ {
		ret = append(ret,uint8(rand.Intn(256)))
	}
	return
}
