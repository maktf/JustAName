package libkademlia

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
	shortlist.active = make(map[string]bool)
	shortList.list = make(map[string]*Contact)
	shortList.target = target
	shortList.removed = make(make[string]bool)
	closestNode = make([]string)
	for _, contact := range localContacts{
		shortList.list[contact.NodeID.AsString()] = contact
		dis := distance(contact.NodeID, target)
		if shortList.shortestDistance < dis{
			shortList.shortestDistance = dis
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
		if dis>shortList.shortestDistance{
			closestNode = make([]string)
			closestNode = append(closestNode,id)
			shortList.shortestDistance = dis
			return true
		}else if dis == shortList.shortestDistance{
			closestNode = append(closestNode, id)
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
	res := make([]*Contact)
	i := 0
	for j:= 0; i<alpha && j<len(shortList.closestNode);j++{
		if _,ok := shortList.calling[shortList.closestNode[j]]; ok == false{
			res = append(res, shortList.list[shortList.closestNode[j]])
			shortList.calling[shortList.closestNode[j]] = true
			i++
		}
	} 
	if(i<alpha){
		for id, contact := range shortList.list{
			_, ok_remove := shortList.removed[id]
			_, ok_calling := shortList.calling[id]
			_, ok_active := shortList.active[id]
			if(ok_active == false && ok_calling == false && ok_remove == false){
				res = append(res, contact)
			}
			i++
			if(i>=alpha)
				break
		}
	}
	return res

}

func (shortList *ShortList) getClosestNodes()[]*Contact{
	return shortList.closestNode;
}

