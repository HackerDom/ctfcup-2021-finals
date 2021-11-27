package commands

import "trash-factory/pkg/serializeb"

type GetItemOp struct {
	ContainerID string
	ItemIndex   int
}

func (op GetItemOp) Serialize() []byte {
	writer := serializeb.NewWriter()
	writer.WriteString(op.ContainerID)
	writer.WriteUint32(op.ItemIndex)
	return writer.GetBytes()
}

func DeserializeGetItemOp(buf []byte) (GetItemOp, error) {
	reader := serializeb.NewReader(buf)
	containerID, err := reader.ReadString()
	if err != nil {
		return GetItemOp{}, err
	}
	itemIndex, err := reader.ReadUint32()
	if err != nil {
		return GetItemOp{}, err
	}

	return GetItemOp{
		ContainerID: containerID,
		ItemIndex:   itemIndex,
	}, nil
}
