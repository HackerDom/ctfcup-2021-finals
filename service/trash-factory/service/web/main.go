package main

import (
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var sessions = NewSessions()

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || r.Method != "GET" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	c, err := r.Cookie("session")
	if err != nil || !sessions.IsCorrectSession(c.Value) {
		uuid, err := sessions.Create("")

		if err != nil {
			log.Error("Can't get session id")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		http.SetCookie(w, &http.Cookie{Name: "session", Value: uuid})
	}


	indexTmpl, err:= template.ParseFiles("templates/index.html")
	if err != nil {
		log.Error("Can't read index.html")
		fmt.Fprintf(w, "Temmplate error. Try again...")
		return
	}
	indexTmpl.Execute(w, nil)

}

type ViewData struct {
	Token string
	TokenKey string
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/new" || r.Method != "GET" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	c, err := r.Cookie("session")
	if err != nil || !sessions.IsCorrectSession(c.Value) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	session := c.Value
	tokenKey, _ := sessions.GetValue(session)
	token := make([]byte, 8)

	if tokenKey == "" {
		binary.BigEndian.PutUint64(token, rand.Uint64())
		tokenKey = fmt.Sprintf("%08x", rand.Uint64())

		err := os.WriteFile("db/users/" + tokenKey,  token, 0666) // TODO: Change token to value
		if err != nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		sessions.UpdateValue(session, tokenKey)
	} else {
		userInfo, err := os.ReadFile("db/users/" + tokenKey)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		token = userInfo[:8]
	}

	indexTmpl, err:= template.ParseFiles("templates/new.html")
	if err != nil {
		log.Error("Can't read new.html")
		fmt.Fprintf(w, "Temmplate error. Try again...")
		return
	}
	indexTmpl.Execute(w, ViewData{TokenKey: tokenKey, Token: fmt.Sprintf("%016x", token)})
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

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/new", newHandler)
	rand.Seed(time.Now().Unix())
	fmt.Printf("Starting server at port :%s\n", port)
	if err := http.ListenAndServe(":" + port, nil); err != nil {
		log.Fatal(err)
	}
}
