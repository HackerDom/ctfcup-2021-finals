package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
	"trash-factory/pkg/api"
)

var (
	sessionsStorage = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	client          = api.NewAdminClient(os.Getenv("CP_ADDR"))
	pageSize        = 20
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Bad request", http.StatusMethodNotAllowed)
		return
	}
	session, err := sessionsStorage.Get(r, "session")
	if err != nil {
		log.Error(err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	if err = session.Save(r, w); err != nil {
		log.Error(err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	indexTmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Error(err)
		http.Error(w, "parse template error", http.StatusInternalServerError)
		return
	}
	indexTmpl.Execute(w, nil)
	if err := indexTmpl.Execute(w, nil); err != nil {
		log.Error(err)
		http.Error(w, "rendering error", http.StatusInternalServerError)
	}

}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Bad request", http.StatusMethodNotAllowed)
		return
	}

	session, err := sessionsStorage.Get(r, "session")
	if err != nil {
		log.Error(err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	if session == nil || session.IsNew {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	tokenKey, found := session.Values["tokenKey"]
	if !found {
		tokenKey, err = client.CreateUser()
		if err != nil {
			log.Error(err)
			http.Error(w, "cant create user", http.StatusBadGateway)
			return
		}
		session.Values["tokenKey"] = tokenKey
		if err = session.Save(r, w); err != nil {
			log.Error(err)
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}
	}

	user, err := client.GetUser(tokenKey.(string))
	if err != nil {
		log.Error(err)
		http.Error(w, "cant get user", http.StatusBadGateway)
		return
	}

	indexTmpl, err := template.ParseFiles("templates/new.html")
	if err != nil {
		log.Error(err)
		http.Error(w, "parse template error", http.StatusInternalServerError)
		return
	}
	if err := indexTmpl.Execute(w, user); err != nil {
		log.Error(err)
		http.Error(w, "rendering error", http.StatusInternalServerError)
	}
}

func statHandler(w http.ResponseWriter, r *http.Request) {
	pageN := 0
	if pageNParam, ok := r.URL.Query()["page"]; ok && len(pageNParam) > 0 {
		pageNParsed, err := strconv.Atoi(pageNParam[0])
		if err == nil && pageNParsed > 0 {
			pageN = pageNParsed - 1
		}
	}

	stat, err := client.GetStat(pageN*pageSize, (pageN+1)*pageSize)
	if err != nil {
		log.Error(err)
		http.Error(w, "cant get stat", http.StatusBadGateway)
		return
	}

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"getPlace": func(i int) int {
			return pageN*pageSize + i + 1
		},
	}

	indexTmpl, err := template.New("stat.html").Funcs(funcMap).ParseFiles("templates/stat.html")
	if err != nil {
		log.Error(err)
		http.Error(w, "parse template error", http.StatusInternalServerError)
		return
	}

	if err := indexTmpl.Execute(w, map[string]interface{}{
		"stat":      stat.Users,
		"prevPageN": pageN,
		"nextPageN": pageN + 2,
	}); err != nil {
		log.Error(err)
		http.Error(w, "rendering error", http.StatusInternalServerError)
		return
	}
}

func main() {
	port, exist := os.LookupEnv("PORT")
	if !exist {
		log.Fatal("PORT not found")
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/stat", statHandler)
	rand.Seed(time.Now().Unix())
	fmt.Printf("Starting server at port :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
