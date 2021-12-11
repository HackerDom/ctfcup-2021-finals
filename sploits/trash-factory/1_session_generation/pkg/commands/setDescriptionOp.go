package commands

import "1_session_generation/pkg/serializeb"

type SetDescriptionOp struct {
	TokenKey    string
	Description string
}

func (op SetDescriptionOp) Serialize() []byte {
	writer := serializeb.NewWriter()
	writer.WriteString(op.TokenKey)
	writer.WriteString(op.Description)
	return writer.GetBytes()
}

func DeserializeSetDescriptionOp(buf []byte) (SetDescriptionOp, error) {
	reader := serializeb.NewReader(buf)
	tokenKey, err := reader.ReadString()
	description, err := reader.ReadString()
	if err != nil {
		return SetDescriptionOp{}, err
	}

	return SetDescriptionOp{
		TokenKey:    tokenKey,
		Description: description,
	}, nil
}
