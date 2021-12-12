package commands

import (
	"1_session_generation/pkg/models"
	"1_session_generation/pkg/serializeb"
)

type PutItemOp struct {
	models.Item
	ContainerId string
}

func (op PutItemOp) Serialize() []byte {
	writer := serializeb.NewWriter()
	op.Item.SerializeNext(&writer)
	writer.WriteString(op.ContainerId)
	return writer.GetBytes()
}

func DeserializePutItemOpOp(buf []byte) (PutItemOp, error) {
	reader := serializeb.NewReader(buf)
	item, err := models.DeserializeNextItem(reader)
	containerId, err := reader.ReadString()
	if err != nil {
		return PutItemOp{}, err
	}
	return PutItemOp{
		item,
		containerId,
	}, nil
}
