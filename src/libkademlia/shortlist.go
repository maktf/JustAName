package libkademlia

type ShortList struct{
	list []*Contact
	flag []int
	count int
	shortestDistance int
}

func (shortList *ShortList) initializeShortList(localContacts []*Contact, target *Contact){
	shortList.shortestDistance = -1
	shortList.count = len(localContacts)
	shortList.list = make([]*Contact, 0, k)
	shortList.flag = make([]int, 0,k)
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
func (shortList *ShortList) updateActiveContact(newContact *Contact) bool{

}


//to remove a certain contact
func (shortList *ShortList) removeInactive(target *Contact) bool{

}

//to check whether either of the two conditions is met
func (shortList *ShortList) checkSucceed() bool{

}

