package commands

import "trash-factory/pkg/models"

type PutItemOp struct {
	models.Item
}

func (op PutItemOp) Serialize() []byte {
	return op.Item.Serialize()
}

func DeserializePutItemOpOp(buf []byte) (PutItemOp, error) {
	item, err := models.DeserializeItem(buf)
	if err != nil {
		return PutItemOp{}, err
	}
	return PutItemOp{
		item,
	}, nil
}
