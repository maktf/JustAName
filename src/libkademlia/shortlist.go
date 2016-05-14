package libkademlia

type ShortList struct{
	active []*Contact
	candlist map(string, *Contact)
	shortestDistance int
	removed map(string, *Contact)
}

func (shortList *ShortList) initializeShortList(localContacts []*Contact, target *Contact){
	shortList.shortestDistance = -1
	shortList.count = len(localContacts)
	shortList.candlist = make(map(string, *Contact))
	for _, contact := range localContacts{
		shortList.list = append(shortList.list, contact)
		dis := distance(contact.NodeID, target.NodeID)
		if shortList.shortestDistance < dis{
			shortList.shortestDistance = dis
		}
		shortList.flag = append(shortList.flag, 0)
	}
}

//To insert a contact into the shortlist, 
//if the contact's distance is larger than the shortest one, the contact will not be added
//false: Unchanged
func (shortList *ShortList) updateActiveContact(newContact *Contact, contactSender *Contact) bool{

}


//to remove a certain contact
func (shortList *ShortList) removeInactive(target *Contact) bool{

}

//to check whether k contacts has been probed and active
func (shortList *ShortList) checkActive() bool{

}

//return alpha contacts which has not been contacted, maybe fewer than alpha
func (shortList *ShortList) getAlphaNotContacted() []*Contact{
	res := make([]*Contact)
	for _,v : range candlist{

	}

}

