package key

import (
	"encoding/binary"
	"errors"
)

func init() {
	RegisterDecoder(V1, Codec{
		Encode: func(k VersionedKey) ([]byte, error) {
			return k.EncodeBinary()
		},
		Decode: decodeV1,
	})
}

type KeyV1 struct {
	NodeID uint64
	DiskID uint64
	Identifier []byte
}

func (k *KeyV1) Version() Version {
	return V1
}

func (k *KeyV1) EncodeBinary() ([]byte, error) {
	identifierLen := uint16(len(k.Identifier))

	bufLen := 4+2*8+identifierLen
	buf := make([]byte, bufLen)

	binary.BigEndian.PutUint16(buf[0:2],   k.Version().Output())
	binary.BigEndian.PutUint16(buf[2:4],   identifierLen)
	binary.BigEndian.PutUint64(buf[4:12],  k.NodeID)
	binary.BigEndian.PutUint64(buf[12:20], k.DiskID)


	copy(buf[20:20+identifierLen], k.Identifier[0:identifierLen])

	return buf, nil
}

func (k *KeyV1) SetIDs(nodeID, diskID uint64) {
	k.NodeID = nodeID
	k.DiskID = diskID
}

func decodeV1(buf []byte) (VersionedKey, error) {
	if len(buf) < 2+8+8 {
		return nil, errors.New("key: invalid V1 key: too short")
	}

	identifierLen := binary.BigEndian.Uint16(buf[0:2])


	nodeId := binary.BigEndian.Uint64(buf[2:10])
	diskId := binary.BigEndian.Uint64(buf[10:18])
	
	idBuf := make([]byte, identifierLen)
	copy(idBuf[0:identifierLen], buf[18:18+identifierLen])

	return &KeyV1{
		NodeID: nodeId,
		DiskID: diskId,
		Identifier: idBuf,
	}, nil
}