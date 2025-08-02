package key

type KeyFactory struct {
	nodeID uint64
	diskId uint64
}

type KeyFactoryParams struct {
	NodeId uint64
	DiskID uint64
}

func NewKeyFactory(params KeyFactoryParams) *KeyFactory {
	return &KeyFactory{
		nodeID: params.NodeId,
		diskId: params.DiskID,
	}
}
