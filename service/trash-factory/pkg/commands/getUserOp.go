package commands

import "trash-factory/pkg/serializeb"

type GetUserOp struct {
	TokenKey string
}

func (op GetUserOp) Serialize() []byte {
	writer := serializeb.NewWriter()
	writer.WriteString(op.TokenKey)
	return writer.GetBytes()
}

func DeserializeGetUserOp(buf []byte) (GetUserOp, error) {
	reader := serializeb.NewReader(buf)
	tokenKey, err := reader.ReadString()
	if err != nil {
		return GetUserOp{}, err
	}

	return GetUserOp{
		TokenKey: tokenKey,
	}, nil
}
