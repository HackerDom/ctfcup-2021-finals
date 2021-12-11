package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
)

var (
	controlPanel = NewControlPanel()
	magic        = []byte{'\x03', '\x13', '\x37'}
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered: ", r)
		}
	}()
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
