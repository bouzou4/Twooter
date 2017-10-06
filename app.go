package main

import (
    "fmt"
    // "io/ioutil"
    "net/http"
    "html/template"
    // "regexp"
    "time"
    "strconv"
    "crypto/sha1"
)

// var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type User struct {
	ID int
	Name string
	Pass string
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

func AddUser(name string, pass string, color string, db *FakeDB) int {
	h := sha1.New()
	h.Write([]byte(pass))
	bs := string(h.Sum(nil))

	tempID := len((*db).Users)
	tempUser := &User{ID: tempID, Name: name, Pass: bs, Color: color, Twoots: []*Twoot{}}
	(*db).Users = append((*db).Users, tempUser)
	return tempID
}

func AddTwoot(author int, body string, db *FakeDB) int {
	tempID := len((*db).Twoots)
	tempTwoot := &Twoot{ID: tempID, Author: (*db).Users[author], Body: body, Created: time.Now()}
	tempTwoots := make([]*Twoot, len((*db).Twoots) + 1)
	tempTwoots[0] = tempTwoot
	copy(tempTwoots[1:], (*db).Twoots)
	(*db).Twoots = tempTwoots
	return tempID
}

func login(username string, password string, db *FakeDB) int {
	for _, usr := range (*db).Users {
		if (*usr).Name == username {
			h := sha1.New()
			h.Write([]byte(password))
			if string(h.Sum(nil)) == (*usr).Pass {
				return (*usr).ID
			}
		}
	}
	return -1
}

//closure that returns a function that takes an http.ResponseWriter and http.Request and includes the FakeDB object
func MakeDbHandler(fn func(http.ResponseWriter, *http.Request, *FakeDB), db *FakeDB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fn(w, r, db)
    }
}

func BaseHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Base GET Request\ncookie value: " + session.Value)
	if session.Value == "" {
		RenderTemplate(w, "login", db)
	} else {
		RenderTemplate(w, "index", db)
	}	
}

func LoginHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	r.ParseForm()
	cookID := login(r.PostFormValue("username"), r.PostFormValue("password"), db)
	fmt.Println("Login Post Request\ncookie value: " + strconv.Itoa(cookID))
	if cookID != -1 {
		tok := http.Cookie {
			Name: "UserID",
			Value: strconv.Itoa(cookID),
			Expires: time.Now().Add(1 * time.Hour),
		}
		http.SetCookie(w, &tok)
	} else {
		tok := http.Cookie {
			Name: "UserID",
			Value: "",
			Expires: time.Now().Add(1 * time.Hour),
		}
		http.SetCookie(w, &tok)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	tok := http.Cookie {
		Name: "UserID",
		Value: "",
	}
	http.SetCookie(w, &tok)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func ComposeHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	switch r.Method {
	case http.MethodGet:
		fmt.Println("Compose GET Request")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	case http.MethodPost:	
		r.ParseForm()
		tok, err := r.Cookie("UserID")
		if err != nil {
	        http.Error(w, err.Error(), http.StatusInternalServerError)
	    }
	    fmt.Println("User: " + tok.Value + " has logged in")
		author,err := strconv.Atoi(tok.Value)
		if err != nil {
	        http.Error(w, err.Error(), http.StatusInternalServerError)
	    }
		AddTwoot(author, r.PostFormValue("twoot"), db)
		fmt.Println("Compose POST Request\nAuthor: " + tok.Value)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func RenderTemplate(w http.ResponseWriter, tmpl string, db *FakeDB) {
	head, err := template.ParseFiles("header.html")
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    content, err := template.ParseFiles(tmpl + ".html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    foot, err := template.ParseFiles("footer.html")
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = head.Execute(w, *db)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    err = content.Execute(w, *db)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    err = foot.Execute(w, *db)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
	db := FakeDB{Users: []*User{}, Twoots: []*Twoot{}}

	AddUser("Adam", "password", "#fae24a", &db)
	AddUser("Rick", "password", "#859911", &db)
	AddUser("Ricardo", "pp", "#a3f5ee", &db)

	AddTwoot(0, "my last name is bouz", &db)
	AddTwoot(0, "what a nice day", &db)
	AddTwoot(0, "whats going on", &db)
	AddTwoot(1, "I like eggs", &db)
	AddTwoot(1, "did you see the game last night", &db)
	AddTwoot(1, "i know who im voting for in the election", &db)
	AddTwoot(2, "any movie recommendations", &db)
	AddTwoot(2, "the last episode of GOT was awesome", &db)
	AddTwoot(2, "check out this hilarious meme", &db)


	http.HandleFunc("/", MakeDbHandler(BaseHandler, &db))
	http.HandleFunc("/login", MakeDbHandler(LoginHandler, &db))
	http.HandleFunc("/logout", MakeDbHandler(LogoutHandler, &db))
	http.HandleFunc("/post", MakeDbHandler(ComposeHandler, &db))
	fmt.Println(http.ListenAndServe(":8080", nil))
}
