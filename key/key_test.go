package key

import (
	"fmt"
	"hash/fnv"
	"testing"
)

func TestKey(t *testing.T) {
	kf := NewKeyFactory(KeyFactoryParams{
		NodeId: hash("node 1"),
		DiskID: hash("disk 1"),
	})

	data, err := kf.EncodeKey(&KeyV1{
		// Data: uint16(57874),
		Identifier: GenerateIdentifier(16),
	})
	if err != nil {
		panic(err)
	}
	

	fmt.Println(data)

	k, err := Decode[*KeyV1](data)
	if err != nil {
		panic(err)
	}

	fmt.Println(k)
}


func hash(data string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(data))
	return h.Sum64()
}