package hashx

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	str1 := "hbyughuyguvhbjgyuhuiigyyu"

	data, err := HashString(FNV64, str1)
	if err != nil {
		panic(err)
	}

	data3, err := HashString(FNV64, str1)
	if err != nil {
		panic(err)
	}

	st := data.Encode()
	fmt.Println(st)

	// h := crc64.New(crc64.MakeTable(crc64.ISO))
	h := sha512.New()

	_, err = h.Write([]byte(str1))
	if err != nil {
		panic(err)
	}

	data2 := h.Sum(nil)

	st2 := hex.EncodeToString(data2)
	fmt.Println(st2)

	fmt.Println(bytes.Compare(data.data, data2))

	fmt.Println("Mod result:", Mod(data.data, 20))
	fmt.Println("Mod result:", QuickMod(data3.data, 20))

	fmt.Println(FNV64.HashString(str1))
}