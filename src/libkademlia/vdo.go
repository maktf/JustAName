package libkademlia

import(
	"time"
	"sss"
	"log"
	"fmt"
)

type VDOReadReq struct {
	key ID
	reschan chan *VanashingDataObject
}

type VDOWriteReq struct {
	key ID
	vdo *VanashingDataObject
}

type VDOManager struct {
	kademlia *Kademlia
	vdoht map[ID]*VanashingDataObject
	readchan chan *VDOReadReq 
	writechan chan *VDOWriteReq
	refreshchan chan ID
	informchan chan int
	epoch int64
	vdotimeset map[ID] int64
}

func (vm *VDOManager) CheckRefresh(){
	for{
		for id, vdo  := range vm.vdoht{
			if(time.Now().Unix()-vm.vdotimeset[id]>vdo.TimeOut){
				//fmt.Println(time.Now().Unix(), vm.vdotimeset[id], vdo.TimeOut)
				vm.refreshchan <- id
				<- vm.informchan
			}
		}
	}
}

func (vm *VDOManager) HandleRequest(k *Kademlia) {
	vm.kademlia = k
	vm.epoch = time.Now().Unix()/28800;
	vm.vdoht = make(map[ID]*VanashingDataObject)
	vm.readchan = make(chan *VDOReadReq)
	vm.writechan = make(chan *VDOWriteReq)
	vm.refreshchan = make(chan ID)
	vm.vdotimeset = make(map[ID] int64)
	vm.informchan = make(chan int)
	for {
		select {
			case r := <- vm.refreshchan:
				vm.epoch = time.Now().Unix()/28800
				res := new(VanashingDataObject)
			     res.AccessKey = vm.vdoht[r].AccessKey
			     res.NumberKeys = vm.vdoht[r].NumberKeys
			     res.Threshold = vm.vdoht[r].Threshold
			     res.Ciphertext = make([]byte, len(vm.vdoht[r].Ciphertext))
			     copy(res.Ciphertext, vm.vdoht[r].Ciphertext)
			    locs := CalculateSharedKeyLocations(res.AccessKey, int64(res.NumberKeys),vm.epoch)
				th := int(res.Threshold)
				i := 0
				keys := make(map[byte][]byte)
				for _, l := range locs {
					_, v, err := vm.kademlia.DoIterativeFindValue(l)
					if err == nil {
						keys[v[0]] = v[1:]
						i++
					}
					if i == th {
						break
					}
				}
				if i == th {
					ckey := sss.Combine(keys)
					keys, err := sss.Split(res.NumberKeys, res.Threshold, ckey)
					if err != nil {
						log.Fatal("NumberKeys or threshold is invalid\n")
					} 
					//akey := GenerateRandomAccessKey()
					//res.AccessKey = akey
					locs = CalculateSharedKeyLocations(res.AccessKey, int64(res.NumberKeys),vm.epoch)
					i = 0
					for key, vs := range keys {
						value := []byte{key}
						for _, v := range vs {
							value = append(value, v)
						}
						_, err := vm.kademlia.DoIterativeStore(locs[i], value)   //what if an error occurs?
						if err != nil {
							log.Fatal("Fail to store shared keys.\n") 
						}
						i++;
					}
					vm.vdotimeset[r] = time.Now().Unix()
					fmt.Println("VDO refresh", r.AsString())
					vm.informchan <- 1
				}

			case r := <- vm.readchan:
			     res := new(VanashingDataObject)
			     res.AccessKey = vm.vdoht[r.key].AccessKey
			     res.NumberKeys = vm.vdoht[r.key].NumberKeys
			     res.Threshold = vm.vdoht[r.key].Threshold
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
		         store.TimeOut = w.vdo.TimeOut
		         vm.vdoht[w.key] = store
		         vm.vdotimeset[w.key] = time.Now().Unix()
		    default:
		}
	}
}

func (vm *VDOManager) Store(key ID, vdo *VanashingDataObject) {
	vm.writechan <- &VDOWriteReq{key, vdo} 
}
