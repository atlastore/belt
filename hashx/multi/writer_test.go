package multi

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/atlastore/belt/hashx"
)

func TestMultiWrite(t *testing.T) {
	r := bytes.NewReader([]byte("gghjkjuytfgbnmkiuytr67uijhgfdr567ujhg"))

	hashers, err := NewHashers(hashx.SHA512, hashx.XXHash, hashx.SHA256, hashx.MD5)
	if err != nil {
		panic(err)
	}

	multiDest := io.MultiWriter(hashers.Writers()...)

	_, err = io.Copy(multiDest, r)
	if err != nil {
		panic(err)
	}

	fmt.Println(hashers.Sum())
}