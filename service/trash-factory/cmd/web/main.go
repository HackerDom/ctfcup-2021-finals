package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	sessionsStorage = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	client   = NewClient("127.0.0.1:9090", "4d65822107fcfd52", "4f163f5f0f9a6278")
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Bad request", http.StatusMethodNotAllowed)
		return
	}
	session, err := sessionsStorage.Get(r, "session")
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = session.Save(r, w); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
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

	session, err := sessionsStorage.Get(r, "session")

	tokenKey, found := session.Values["tokenKey"]
	if !found {
		tokenKey, err = client.createUser()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["tokenKey"] = tokenKey
		if err = session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	user, err := client.getUser(tokenKey.(string))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indexTmpl, err := template.ParseFiles("templates/new.html")
	if err != nil {
		log.Error("Can't read new.html")
		return
	}
	indexTmpl.Execute(w, user)
}

func statHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	port, exist := os.LookupEnv("PORT")
	if !exist {
		log.Fatal("PORT not found")
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/new", newHandler)
	http.HandleFunc("/stat", newHandler)
	rand.Seed(time.Now().Unix())
	fmt.Printf("Starting server at port :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func main1() {
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

	stat, err := client.getStat(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	for i, user := range stat.Users {
		log.Infof("#%d   %x -- %d", i, user.TokenKey, user.Total)
	}
}
