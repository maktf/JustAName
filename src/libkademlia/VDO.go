package libkademlia

type VDOReadReq struct {
	key ID
	reschan chan *VanashingDataObject
}

type VDOWriteReq struct {
	key ID
	vdo *VanashingDataObject
}

type VDOManager struct {
	vdoht map[ID]*VanashingDataObject
	readchan chan *VDOReadReq 
	writechan chan *VDOWriteReq
}

func (vm *VDOManager) HandleRequest() {
	vm.vdoht = make(map[ID]*VanashingDataObject)
	vm.readchan = make(chan *VDOReadReq)
	vm.writechan = make(chan *VDOWriteReq)
	for {
		select {
			case r := <- vm.readchan:
			     res := new(VanashingDataObject)
			     res.AccessKey = vm.vdoht[r.key].AccessKey
			     res.NumberKeys = vm.vdoht[r.key].NumberKeys
			     res.NumberKeys = vm.vdoht[r.key].Threshold
			     res.Ciphertext = make([]byte, len(vm.vdoht[r.key].Ciphertext))
			     copy(res.Ciphertext, vm.vdoht[r.key].Ciphertext)
			     r.reschan <- res
		    case w := <- vm.writechan:
		         store := new(VanashingDataObject)
		         store.AccessKey = w.vdo.AccessKey
		         store.NumberKeys = w.vdo.NumberKeys
		         store.Threshold = w.vdo.Threshold
		         store.Ciphertext = make([]byte, len(w.vdo.Ciphertext))
		         copy(store.Ciphertext, w.vdo.Ciphertext)
		         vm.vdoht[w.key] = store
		    default:
		}
	}
}

func (vm *VDOManager) Store(key ID, vdo *VanashingDataObject) {
	vm.writechan <- &VDOWriteReq{key, vdo} 
}
