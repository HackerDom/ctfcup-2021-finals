package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"os"
	"strings"
	"trash-factory/pkg/commands"
	"trash-factory/pkg/models"
)

var (
	controlPanel = NewControlPanel()
	magic        = []byte{'\x03', '\x13', '\x37'}
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	_, err := conn.Write(magic)
	if err != nil {
		log.Error(err)
		return
	}

	addr, err := net.LookupHost(os.Getenv("WEB_ADDR"))
	if err != nil {
		panic(fmt.Sprintf("%s. %s", "Can't resolve backend url", err.Error()))
	}
	remoteAddr := strings.Split(conn.RemoteAddr().String(), ":")[0]
	if addr[0] == remoteAddr {
		write, err := conn.Write(controlPanel.AdminCredentials.Serialize())
		if err != nil {
			log.Error(write)
			return
		}
	}

	buffer := make([]byte, 128)
	count, err := conn.Read(buffer)

	if err != nil || count < 1 {
		return
	}

	fmt.Printf("Got %d\n", count)
	fmt.Println(buffer)

	statusCode, response := controlPanel.ProcessMessage(buffer[:count])
	if response != nil {
		conn.Write(append([]byte{statusCode}, response...))
	} else {
		conn.Write([]byte{statusCode})
	}
}

func main() {
	err := GenerateTestData()
	if err != nil {
		return
	}

	port, exist := os.LookupEnv("PORT")
	if !exist {
		log.Fatal("PORT not found")
	}

	if _, err := os.Stat("db/users"); os.IsNotExist(err) {
		log.Fatal("Folder db/users not exist")
	}

	if _, err := os.Stat("db/containers"); os.IsNotExist(err) {
		log.Fatal("Folder db/containers not exist")
	}

	l, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	log.Infof("Server started on :%s", port)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		go handleConn(conn)
	}

}

func GenerateTestData() error {
	for i := 0; i < 3; i++ {
		tokenKey, err := AddTestUser()
		if err != nil {
			return err
		}
		err = AddTestContainer(tokenKey)
		if err != nil {
			return err
		}
		err = AddTestContainer(tokenKey)
		if err != nil {
			return err
		}
		for j := 0; j < i+1; j++ {
			err = PutItem(tokenKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func PutItem(tokenKey string) error {
	putItemOp := commands.PutItemOp{
		models.Item{
			Type:        1,
			Description: "trash" + fmt.Sprintf("%08x", rand.Uint64()),
			Weight:      10,
		},
	}
	_, err := controlPanel.PutItem(tokenKey, putItemOp.Serialize())
	if err != nil {
		log.Error(err)
		return err
	}

	user, err := controlPanel.GetUser(tokenKey, commands.GetUserOp{
		TokenKey: tokenKey,
	}.Serialize())
	if err != nil {
		log.Error(err)
		return err
	}

	deserializeUser, err := models.DeserializeUser(user)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, id := range deserializeUser.ContainersIds {
		info, err := controlPanel.GetContainerInfo(tokenKey, commands.GetContainerInfoOp{ContainerID: id}.Serialize())
		if err != nil {
			log.Error(err)
			return err
		}

		container, err := models.DeserializeContainer(info)
		if err != nil {
			log.Error(err)
			return err
		}
		for _, item := range container.Items {
			if item.Weight == putItemOp.Weight &&
				item.Type == putItemOp.Type &&
				item.Description == putItemOp.Description {
				return nil
			}
		}

	}
	log.Error("Item not found")
	return errors.New("Item not found")
}

func AddTestContainer(tokenKey string) error {
	containerOp := commands.CreateContainerOp{
		Size:        5,
		Description: "Fill me up senpai" + fmt.Sprintf("%08x", rand.Uint64()),
	}
	_, err := controlPanel.CreateContainer(tokenKey, containerOp.Serialize())
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func AddTestUser() (string, error) {
	tokenKey := fmt.Sprintf("%08x", rand.Uint64())
	t := make([]byte, 8)
	binary.LittleEndian.PutUint64(t, rand.Uint64())
	op := commands.CreateUserOp{
		Token:    t,
		TokenKey: tokenKey,
	}
	_, err := controlPanel.CreateUser(controlPanel.AdminCredentials.TokenKey, op.Serialize())
	if err != nil {
		log.Error(err)
		return "", err
	}

	user, err := controlPanel.GetUser(controlPanel.AdminCredentials.TokenKey, commands.GetUserOp{
		TokenKey: tokenKey,
	}.Serialize())
	if err != nil {
		return "", err
	}

	deserializeUser, err := models.DeserializeUser(user)
	if err != nil {
		return "", err
	}

	if deserializeUser.TokenKey != tokenKey || string(deserializeUser.Token) != string(op.Token) {
		log.Error("User serialization bug", user, deserializeUser)
	}
	return tokenKey, err
}
