package key

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
)

func Decode[T VersionedKey](encoded string) (T, error) {
	var zero T
	raw, err := hex.DecodeString(encoded)
	if err != nil {
		return zero, fmt.Errorf("key: failed to hex decode: %v", err)
	}

	if len(raw) < 2 {
		return zero, fmt.Errorf("key: too short")
	}

	version := Version(binary.BigEndian.Uint16(raw[:2]))
	payload := raw[2:]

	registryMu.RLock()
	codec, ok := codecRegistry[version]
	registryMu.RUnlock()
	
	if !ok {
		return zero, fmt.Errorf("key: no decoder registered for version %d", version)
	}

	key, err := codec.Decode(payload)
	if err != nil {
		return zero, err
	}

	expectedType := reflect.TypeOf(zero)

	if reflect.TypeOf(key) != expectedType {
		return zero, fmt.Errorf("key: decoded key is not of excepted type %T", zero)
	}

	typedKey := key.(T)

	return typedKey, nil
}