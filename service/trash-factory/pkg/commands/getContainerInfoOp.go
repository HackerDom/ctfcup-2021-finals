package commands

import "trash-factory/pkg/serializeb"

type GetContainerInfoOp struct {
	ContainerID string
}

func (op GetContainerInfoOp) Serialize() ([]byte, error) {
	writer := serializeb.NewWriter()
	writer.WriteString(op.ContainerID)
	return writer.GetBytes()
}

func DeserializeGetContainerInfoOp(buf []byte) (GetContainerInfoOp, error) {
	reader := serializeb.NewReader(buf)
	id, err := reader.ReadString()
	if err != nil {
		return GetContainerInfoOp{}, err
	}

	return GetContainerInfoOp{
		ContainerID: id,
	}, nil
}
