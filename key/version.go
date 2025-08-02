package key


type Version uint16

const (
	V1 Version = iota
)

const CurrVersion = V1

func (v Version) Output() uint16 {
	return uint16(v)
}

type VersionedKey interface {
	EncodeBinary() ([]byte, error)
	Version() Version
	SetIDs(nodeID, diskID uint64)
}