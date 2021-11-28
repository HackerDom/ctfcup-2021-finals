package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	sessions = NewSessions()
	client   = NewClient("127.0.0.1:9090", "4d65822107fcfd52", "4f163f5f0f9a6278")
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Bad request", http.StatusMethodNotAllowed)
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

	indexTmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Error("Can't read index.html")
		return
	}
	indexTmpl.Execute(w, nil)

}

func newHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Bad request", http.StatusMethodNotAllowed)
		return
	}

	c, err := r.Cookie("session")
	if err != nil || !sessions.IsCorrectSession(c.Value) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	session := c.Value
	tokenKey, _ := sessions.GetValue(session)

	if tokenKey == "" {
		tokenKey, err := client.createUser()
		if err != nil {
			log.Error(err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		sessions.UpdateValue(session, tokenKey)
	}

	user, err := client.getUser(tokenKey)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	indexTmpl, err := template.ParseFiles("templates/new.html")
	if err != nil {
		log.Error("Can't read new.html")
		return
	}
	indexTmpl.Execute(w, user)
}

func main1() {
	port, exist := os.LookupEnv("PORT")
	if !exist {
		log.Fatal("PORT not found")
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/new", newHandler)
	rand.Seed(time.Now().Unix())
	fmt.Printf("Starting server at port :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	tokenKey, err := client.createUser()
	if err != nil {
		log.Fatal(err)
	}
	rand.Seed(time.Now().Unix())
	log.Info(tokenKey)
	user, err := client.getUser(tokenKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(user)
}
