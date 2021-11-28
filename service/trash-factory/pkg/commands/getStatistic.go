package commands

import "trash-factory/pkg/serializeb"

type GetStatisticOp struct {
	Skip int
	Take int
}

func (op GetStatisticOp) Serialize() []byte {
	writer := serializeb.NewWriter()
	writer.WriteUint32(op.Skip)
	writer.WriteUint32(op.Take)
	return writer.GetBytes()
}

func DeserializeGetStatisticOp(buf []byte) (GetStatisticOp, error) {
	reader := serializeb.NewReader(buf)
	skip, err := reader.ReadUint32()
	take, err := reader.ReadUint32()
	if err != nil {
		return GetStatisticOp{}, err
	}

	return GetStatisticOp{
		Skip: skip,
		Take: take,
	}, nil
}
