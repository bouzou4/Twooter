package main

import (
    "fmt"
    // "io/ioutil"
    "net/http"
    "html/template"
    // "regexp"
    "time"
)

// var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type User struct {
	ID int
	Name string
	Color string
	Twoots []*Twoot
}

type Twoot struct {
	ID int
	Author *User
	Body string
	Created time.Time
}

type FakeDB struct {
	Users []*User
	Twoots []*Twoot
}

//closure that returns a function that takes an http.ResponseWriter and http.Request and includes the FakeDB object
func MakeDbHandler(fn func(http.ResponseWriter, *http.Request, *FakeDB), db *FakeDB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fn(w, r, db)
    }
}

func BaseHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	RenderTemplate(w, "index", db)
}

func RenderTemplate(w http.ResponseWriter, tmpl string, db *FakeDB) {
    t, err := template.ParseFiles(tmpl + ".html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, *db)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
	db := FakeDB{Users: []*User{}, Twoots: []*Twoot{}}

	u1 := &User{ID: len(db.Users), Name: "Adam", Color: "#fae24a", Twoots: []*Twoot{}}
	u2 := &User{ID: len(db.Users), Name: "Rick", Color: "#859911", Twoots: []*Twoot{}}
	u3 := &User{ID: len(db.Users), Name: "Ricardo", Color: "#a3f5ee", Twoots: []*Twoot{}}

	t1 := &Twoot{ID: len(db.Twoots), Author: u1, Body: "my last name is bouz", Created: time.Now()}
	t2 := &Twoot{ID: len(db.Twoots), Author: u1, Body: "what a nice day", Created: time.Now()}
	t3 := &Twoot{ID: len(db.Twoots), Author: u1, Body: "whats going on", Created: time.Now()}
	t4 := &Twoot{ID: len(db.Twoots), Author: u2, Body: "I like eggs", Created: time.Now()}
	t5 := &Twoot{ID: len(db.Twoots), Author: u2, Body: "did you see the game last night", Created: time.Now()}
	t6 := &Twoot{ID: len(db.Twoots), Author: u2, Body: "i know who im voting for in the election", Created: time.Now()}
	t7 := &Twoot{ID: len(db.Twoots), Author: u3, Body: "any movie recommendations", Created: time.Now()}
	t8 := &Twoot{ID: len(db.Twoots), Author: u3, Body: "the last episode of GOT was awesome", Created: time.Now()}
	t9 := &Twoot{ID: len(db.Twoots), Author: u3, Body: "check out this hilarious meme", Created: time.Now()}
	(*u1).Twoots = append((*u1).Twoots, t1)
	(*u1).Twoots = append((*u1).Twoots, t2)
	(*u1).Twoots = append((*u1).Twoots, t3)
	(*u2).Twoots = append((*u2).Twoots, t4)
	(*u2).Twoots = append((*u2).Twoots, t5)
	(*u2).Twoots = append((*u2).Twoots, t6)
	(*u3).Twoots = append((*u3).Twoots, t7)
	(*u3).Twoots = append((*u3).Twoots, t8)
	(*u3).Twoots = append((*u3).Twoots, t9)
	db.Users = append(db.Users, u1)
	db.Users = append(db.Users, u2)
	db.Users = append(db.Users, u3)
	db.Twoots = append(db.Twoots, t1)
	db.Twoots = append(db.Twoots, t2)
	db.Twoots = append(db.Twoots, t3)
	db.Twoots = append(db.Twoots, t4)
	db.Twoots = append(db.Twoots, t5)
	db.Twoots = append(db.Twoots, t6)
	db.Twoots = append(db.Twoots, t7)
	db.Twoots = append(db.Twoots, t8)
	db.Twoots = append(db.Twoots, t9)


	http.HandleFunc("/", MakeDbHandler(BaseHandler, &db))
	fmt.Println(http.ListenAndServe(":8080", nil))
}
