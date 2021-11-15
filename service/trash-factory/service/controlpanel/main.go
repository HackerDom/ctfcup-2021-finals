package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
)

var (
	controlPanel = NewControlPanel()
	greeting =  []byte {'\x03','\x13','\x37'}
)

func handleConn(conn net.Conn) {
	_, err := conn.Write(greeting)
	if err != nil {
		log.Error(err)
		return
	}

	buffer := make([]byte, 128)
	for {
		count, err := conn.Read(buffer)

		if err != nil || count < 1 {
			return
		}

		fmt.Printf("Got %d\n", count)
		fmt.Println(buffer)

		statusCode := controlPanel.ProcessMessage(buffer[:count])
		conn.Write([]byte {statusCode})
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

	l, err := net.Listen("tcp4", ":" + port)
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
