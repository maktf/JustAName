package libkademlia

import(
	"fmt"
)

type ShortList struct{
	active map[string]bool
	list map[string]*Contact
	calling map[string]bool
	shortestDistance int
	closestNode []string
	removed map[string]bool
	target ID
}

func (shortList *ShortList) initializeShortList(localContacts []*Contact, target ID){
	shortList.shortestDistance = -1
	shortList.calling = make(map[string]bool)
	shortList.active = make(map[string]bool)
	shortList.list = make(map[string]*Contact)
	shortList.target = target
	shortList.removed = make(map[string]bool)
	shortList.closestNode = make([]string,0)
	for _, contact := range localContacts{
		shortList.list[contact.NodeID.AsString()] = contact
		dis := distance(contact.NodeID, target)
		if shortList.shortestDistance < dis{
			shortList.shortestDistance = dis
			shortList.closestNode = append(shortList.closestNode,contact.NodeID.AsString())
		}else if shortList.shortestDistance == dis{
			shortList.closestNode = append(shortList.closestNode,contact.NodeID.AsString())
		}
	}
}

//To insert a contact into the shortlist, 
//if the closestNode is changed, return true. Else, return false
func (shortList *ShortList) updateActiveContact(newContact *Contact) bool{
	dis := distance(newContact.NodeID, shortList.target)
	id := newContact.NodeID.AsString()
	if _,ok := shortList.list[id]; ok == false{
		shortList.list[id] = newContact
		fmt.Println("Insert: ", id, " ", newContact.NodeID.AsString())
		if dis>shortList.shortestDistance{
			shortList.closestNode = make([]string,0)
			shortList.closestNode = append(shortList.closestNode,id)
			shortList.shortestDistance = dis
			return true
		}else if dis == shortList.shortestDistance{
			shortList.closestNode = append(shortList.closestNode, id)
			return true
		}
	}
	return false
}

func (shortList *ShortList) setActive(target *Contact){
	id := target.NodeID.AsString()
	if len(shortList.active) < IDBytes{
		shortList.active[id] = true
	}
	delete(shortList.calling, id)
}


//to remove a certain contact
func (shortList *ShortList) removeInactive(target *Contact) bool{
	id := target.NodeID.AsString()
	shortList.removed[target.NodeID.AsString()] = true
	delete(shortList.calling, id)
	return true
}

//to check whether k contacts has been probed and active
func (shortList *ShortList) checkActive() bool{
	if len(shortList.active) == k{
		return true
	}
	return false
}

//return alpha contacts which has not been contacted, maybe fewer than alpha
func (shortList *ShortList) getAlphaNotContacted() []*Contact{
	res := make([]*Contact,0)
	i := 0
	for j:= 0; i<alpha && j<len(shortList.closestNode);j++{
		if _,ok := shortList.calling[shortList.closestNode[j]]; ok == false{
			res = append(res, shortList.list[shortList.closestNode[j]])
			fmt.Println("getAlpha: ", shortList.closestNode[j], " ", shortList.list[shortList.closestNode[j]].NodeID.AsString())
			shortList.calling[shortList.closestNode[j]] = true
			i++
		}
	} 
	if i<alpha {
		for id, contact := range shortList.list {
			_, ok_remove := shortList.removed[id]
			_, ok_calling := shortList.calling[id]
			_, ok_active := shortList.active[id]
			if ok_active == false && ok_calling == false && ok_remove == false {
				res = append(res, contact)
				fmt.Println("<alpha: ", contact.NodeID.AsString())
				i++
			}
			
			if i>=alpha {
				break
			}
		}
	}
	
	for _, r := range res {
	     fmt.Println("result: ", (*r).NodeID.AsString())
	}
	 
	return res

}

func (shortList *ShortList) getActiveNodes() []*Contact{
	res := make([]*Contact,0)
	for id := range shortList.active{
		res = append(res, shortList.list[id])
	}
	return res
}

func (shortList *ShortList) getClosestNodes()[]*Contact{
	res := make([]*Contact,0)
	for _,id := range shortList.closestNode{
		res = append(res, shortList.list[id])
	}
	return res
}

func (shortList *ShortList) getAllNotContacted() []*Contact{
	res := make([]*Contact,0)
	for id, contact := range shortList.list{
		_, ok_remove := shortList.removed[id]
		_, ok_calling := shortList.calling[id]
		_, ok_active := shortList.active[id]
		if(ok_active == false && ok_calling == false && ok_remove == false){
			res = append(res, contact)
		}
	}
	return res
}

func (shortList *ShortList) printStatus() {
	for str := range shortList.active {
		fmt.Println("Print Active: ", str)
	}
	for str := range shortList.removed {
		fmt.Println("Print Removed: ", str)
	}
	for str := range shortList.calling {
		fmt.Println("Print Calling: ", str)
	}
	for _, str := range shortList.closestNode {
		fmt.Println("Print Closest: ", str)
	}
	for str, v := range shortList.list {
		fmt.Println("Print List: ", str, v.NodeID.AsString())
	}
}
