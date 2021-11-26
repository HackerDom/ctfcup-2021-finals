package commands

import "trash-factory/pkg/serializeb"

type CreateUserOp struct {
	Token []byte
}

func (op CreateUserOp) Serialize() ([]byte, error) {
	writer := serializeb.NewWriter()
	writer.WriteBytes(op.Token)
	return writer.GetBytes()
}

func DeserializeCreateUserOp(buf []byte) (CreateUserOp, error) {
	reader := serializeb.NewReader(buf)
	token, err := reader.ReadBytes()
	if err != nil {
		return CreateUserOp{}, err
	}

	return CreateUserOp{
		Token: token,
	}, nil
}
