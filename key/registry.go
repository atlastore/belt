package key

import (
	"fmt"
	"sync"
)

var (
	codecRegistry = make(map[Version]Codec)
	registryMu      sync.RWMutex
)

type Codec struct {
	Encode func(VersionedKey) ([]byte, error)
	Decode func([]byte) (VersionedKey, error)
}

func RegisterDecoder(version Version, codec Codec) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := codecRegistry[version]; exists {
		panic(fmt.Sprintf("decoder for version %d already registered", version))
	}
	codecRegistry[version] = codec
}