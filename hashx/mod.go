package hashx

import (
	"encoding/binary"
	"math/big"
)


func Mod(data []byte, mod int64) int64 {
	hashInt := new(big.Int).SetBytes(data[:])

	modVal := new(big.Int).Mod(hashInt, big.NewInt(mod))

	return modVal.Int64()
}

func QuickMod(data []byte, mod int64) int64 {
	buf := make([]byte, 8)
	if len(data) > 8 {
		copy(buf, data[:8])
	} else {
		copy(buf[8-len(data):], data) // pad on the left
	}

	hashInt := binary.BigEndian.Uint64(buf)

	modVal := hashInt % uint64(mod)

	return int64(modVal)
}