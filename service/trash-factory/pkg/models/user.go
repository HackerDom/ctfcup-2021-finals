package models

import (
	"trash-factory/pkg/serializeb"
)

type User struct {
	TokenKey      string
	Token         []byte
	ContainersIds []string
}

func (user *User) Serialize() ([]byte, error) {
	writer := serializeb.NewWriter()
	user.SerializeNext(&writer)
	return writer.GetBytes()
}

func (user *User) SerializeNext(writer *serializeb.Writer) {
	writer.WriteString(user.TokenKey)
	writer.WriteBytes(user.Token)
	writer.WriteArray(serializeb.ToGenericArray(user.ContainersIds),
		func(item interface{}, writer *serializeb.Writer) {
			writer.WriteString(item.(string))
		})
}

func DeserializeUser(buf []byte) (User, error) {
	reader := serializeb.NewReader(buf)
	return DeserializeNextUser(reader)
}

func DeserializeNextUser(reader serializeb.Reader) (User, error) {
	tokenKey, err := reader.ReadString()
	if err != nil {
		return User{}, err
	}
	token, err := reader.ReadBytes()
	if err != nil {
		return User{}, err
	}

	size, err := reader.ReadArraySize()
	if err != nil {
		return User{}, err
	}

	containersIds := make([]string, size)
	for i := 0; i < size; i++ {
		containersIds[i], err = reader.ReadString()
		if err != nil {
			return User{}, err
		}
	}

	return User{
		TokenKey:      tokenKey,
		Token:         token,
		ContainersIds: containersIds,
	}, nil
}
