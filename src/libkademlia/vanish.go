package libkademlia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	mathrand "math/rand"
	"time"
	"sss"
	"log"
)

type VanashingDataObject struct {
	AccessKey  int64
	Ciphertext []byte
	NumberKeys byte
	Threshold  byte
	TimeOut int64
}

func GenerateRandomCryptoKey() (ret []byte) {
	for i := 0; i < 32; i++ {
		ret = append(ret, uint8(mathrand.Intn(256)))
	}
	return
}

func GenerateRandomAccessKey() (accessKey int64) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	accessKey = r.Int63()
	return
}

func CalculateSharedKeyLocations(accessKey int64, count int64, epoch int64) (ids []ID) {
	r := mathrand.New(mathrand.NewSource(accessKey+epoch))
	ids = make([]ID, count)
	for i := int64(0); i < count; i++ {
		for j := 0; j < IDBytes; j++ {
			ids[i][j] = uint8(r.Intn(256))
		}
	}
	return
}

func encrypt(key []byte, text []byte) (ciphertext []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext = make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return
}

func decrypt(key []byte, ciphertext []byte) (text []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext is not long enough")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext
}

func (k *Kademlia) VanishData(data []byte, numberKeys byte,
	threshold byte, timeoutSeconds int) (vdo VanashingDataObject) {
	ckey := GenerateRandomCryptoKey()
	vdo.Ciphertext = encrypt(ckey, data)
	vdo.NumberKeys = numberKeys
	vdo.Threshold = threshold
	keys, err := sss.Split(numberKeys, threshold, ckey)
	if err != nil {
		log.Fatal("NumberKeys or threshold is invalid\n")
	} 
	
	akey := GenerateRandomAccessKey()
	vdo.AccessKey = akey
	locs := CalculateSharedKeyLocations(akey, int64(numberKeys), k.VM.epoch)
	i := 0
	for key, vs := range keys {
		value := []byte{key}
		for _, v := range vs {
			value = append(value, v)
		}
		_, err := k.DoIterativeStore(locs[i], value)   //what if an error occurs?
		if err != nil {
			log.Fatal("Fail to store shared keys.\n") 
		}
		i++;
	}
	vdo.TimeOut = int64(timeoutSeconds)
	return
}

func (k *Kademlia) UnvanishData(vdo VanashingDataObject) (data []byte) {
	locs := CalculateSharedKeyLocations(vdo.AccessKey, int64(vdo.NumberKeys), k.VM.epoch)
	th := int(vdo.Threshold)
	i := 0
	keys := make(map[byte][]byte)
	for _, l := range locs {
		_, v, err := k.DoIterativeFindValue(l)
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
		return decrypt(ckey, vdo.Ciphertext)
	} else {
		return nil
	}
}
