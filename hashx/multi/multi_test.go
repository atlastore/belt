package multi

import (
	"fmt"
	"testing"

	"github.com/atlastore/belt/hashx"
)


func TestMulti(t *testing.T) {
	str1 := "gghjkjuytfgbnmkiuytr67uijhgfdr567ujhg"
	hashes, err := HashString(str1, hashx.SHA512, hashx.XXHash, hashx.SHA256, hashx.MD5)
	if err != nil {
		panic(err)
	}

	fmt.Println(hashes)
}