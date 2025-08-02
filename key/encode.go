package key

import (
	"encoding/hex"
	"fmt"
)

func (kf *KeyFactory) EncodeKey(k VersionedKey) (string, error) {
	k.SetIDs(kf.nodeID, kf.diskId)

	registryMu.RLock()
	codec, ok := codecRegistry[k.Version()]
	registryMu.RUnlock()
	if !ok {
		return "", fmt.Errorf("key: no encoder registered for version %d", k.Version())
	}

	body, err := codec.Encode(k)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(body), nil
}