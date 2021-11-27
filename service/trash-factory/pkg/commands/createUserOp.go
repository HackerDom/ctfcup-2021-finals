package commands

import "trash-factory/pkg/serializeb"

type CreateUserOp struct {
	Token    []byte
	TokenKey string
}

func (op CreateUserOp) Serialize() []byte {
	writer := serializeb.NewWriter()
	writer.WriteBytes(op.Token)
	writer.WriteString(op.TokenKey)
	return writer.GetBytes()
}

func DeserializeCreateUserOp(buf []byte) (CreateUserOp, error) {
	reader := serializeb.NewReader(buf)
	token, err := reader.ReadBytes()
	tokenKey, err := reader.ReadString()
	if err != nil {
		return CreateUserOp{}, err
	}

	return CreateUserOp{
		Token:    token,
		TokenKey: tokenKey,
	}, nil
}
