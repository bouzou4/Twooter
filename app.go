//	I apologize in advance for this mess
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

//	regex to validate url request to be implemented soon(tm) 
//	var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

//	basic user struct with list of pointers to their created Twoots
type User struct {
	ID int
	Name string
	Pass string
	Color string
	FollowList []*User
	FollowedList []*User
	Twoots []*Twoot
}

type Twoot struct {
	ID int
	Author *User
	Body string
	Created time.Time
}

//	no files yet so the database is held in memory meaning memory violations are always a hair away
type FakeDB struct {
	Users []*User
	Twoots []*Twoot
}

type Instance struct {
	Client *User
	Timeline []*Twoot
	DB *FakeDB
}


//	adds a user to the database while storing their password as a hash
//	returns UserID
func AddUser(name string, pass string, color string, db *FakeDB) int {
	h := sha1.New()
	h.Write([]byte(pass))
	bs := string(h.Sum(nil))

	tempID := len((*db).Users)
	tempUser := &User{
					ID: tempID, 
					Name: name, 
					Pass: bs, 
					Color: color, 
					FollowList: []*User{},
					Twoots: []*Twoot{},
				}
	(*db).Users = append((*db).Users, tempUser)
	return tempID
}

//	creates and adds Twoot to the database
//	returns TwootID
func AddTwoot(author int, body string, db *FakeDB) int {
	tempID := len((*db).Twoots)
	tempAuth := (*db).Users[author]
	tempTwoot := &Twoot{
					ID: tempID, 
					Author: tempAuth, 
					Body: body, 
					Created: time.Now(),
				}
	
	tempTwoots := make([]*Twoot, len((*db).Twoots) + 1)
	tempTwoots[0] = tempTwoot
	copy(tempTwoots[1:], (*db).Twoots)
	(*db).Twoots = tempTwoots

	tempTwoots = make([]*Twoot, len((*tempAuth).Twoots) + 1)
	tempTwoots[0] = tempTwoot
	copy(tempTwoots[1:], (*tempAuth).Twoots)
	(*tempAuth).Twoots = tempTwoots

	return tempID
}

func Follow(user int, following int, db *FakeDB) {
	db.Users[user].FollowList = append(db.Users[user].FollowList, db.Users[following])
	db.Users[following].FollowedList = append(db.Users[following].FollowedList, db.Users[user])
}

//	gets index of followed account and returns it or -1 if not found
func TwootIndex(vs []*User, t *User) int {
    for i, v := range vs {
        if (*v).ID == (*t).ID {
            return i
        }
    }
    return -1
}

//	filters out Twoots in database to find those asked for by follow list
//	returns list of pointers to them
func FollowFilter(follows []*User, db *FakeDB) []*Twoot {
	timeline := []*Twoot{}
	for _, i := range (*db).Twoots {
		if TwootIndex(follows, (*i).Author) != -1 {
			timeline = append(timeline, i)
		}
	}
	return timeline
}

//	resets all IDs in a list of Twoots to their proper order
func SortTwoots(list *[]*Twoot) {
	for i,x := range *list {
		(*x).ID = len(*list) - i
	}
	fmt.Println("sorted twoots")
}

//	Used to remove a Twoot from the DB given it's ID
func DeleteTwoot(dID int, db *FakeDB) {
	fmt.Printf("looking for Twoot: dID")
	for x := range (*db).Twoots {
		if (*(*db).Twoots[x]).ID == dID {
			fmt.Print("deleting twoot: ")
			fmt.Print((*(*db).Twoots[x]))
			fmt.Print("\n")
			copy((*db).Twoots[x:], (*db).Twoots[x + 1:])
			(*db).Twoots[len((*db).Twoots) - 1] = nil
			(*db).Twoots = (*db).Twoots[:len((*db).Twoots) - 1]
			SortTwoots(&db.Twoots)
			break
		}
	}
}

//	Used to remove a User from the DB given their ID
func DeleteUser(delID int, db *FakeDB) {
	for x := range (*db).Users {
		if (*(*db).Users[x]).ID == delID {
			fmt.Printf("deleting user: %s\n", (*(*db).Users[x]).Name)
			for _,y := range (*(*db).Users[x]).Twoots {
				DeleteTwoot((*y).ID, db)
			}
			copy((*db).Users[x:], (*db).Users[x + 1:])
			(*db).Users[len((*db).Users) - 1] = nil
			(*db).Users = (*db).Users[:len((*db).Users) - 1]
			break
		}
	}
}

//	function used to get userID from Username
//	returns -1 if not found
func GetID(username string, db *FakeDB) int {
	for _, usr := range db.Users {
		if usr.Name == username {
			return usr.ID
		}
	}
	return -1
}

//	function used to check if username and password hash match up
//	returns -1 if credentials are invalid
func login(username string, password string, db *FakeDB) int {
	uID := GetID(username, db)
	h := sha1.New()
	h.Write([]byte(password))
	if string(h.Sum(nil)) == db.Users[uID].Pass {
		return uID
	}
	return -1
}

//	closure that returns a function that takes an http.ResponseWriter and http.Request and includes the FakeDB object
func MakeDbHandler(fn func(http.ResponseWriter, *http.Request, *FakeDB), db *FakeDB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fn(w, r, db)
    }
}


//	webhandler for the homepage, if the user is logged in then they get their timeline 
//	otherwise they get the login page
func BaseHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
		RenderFileTemplate(w, "login", db)
	} else {
		if session.Value == "" {
			RenderFileTemplate(w, "login", db)
		} else {
			RenderTimeline(w, r, db)
		}
	}
}

//	webhandler for login page performs login() and sets the cookie
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
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	} else {
		tok := http.Cookie {
			Name: "UserID",
			Value: "",
		}
		http.SetCookie(w, &tok)
		http.Redirect(w, r, "/loginfail", http.StatusTemporaryRedirect)
	}
}

//	webhandler for displaying the failed login page; redirects after 5 seconds
func LoginFailHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	RenderFileTemplate(w, "loginfail", db)
}

//	webhandler for logout; essentially just clears cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	tok := http.Cookie {
		Name: "UserID",
		Value: "",
	}
	http.SetCookie(w, &tok)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//	webhandler for posting Twoots, will not post if text is longer than 100 chars
func ComposeHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	switch r.Method {
	case http.MethodGet:
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	case http.MethodPost:	
		r.ParseForm()
		if len(r.PostFormValue("twoot")) <= 100 {
			tok, err := r.Cookie("UserID")
			if err != nil {
		        http.Error(w, err.Error(), http.StatusInternalServerError)
		    }
			author,err := strconv.Atoi(tok.Value)
			if err != nil {
		        http.Error(w, err.Error(), http.StatusInternalServerError)
		    }
			AddTwoot(author, r.PostFormValue("twoot"), db)
		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

//	webhandler for registering users and detecting if username is already in use
func RegisterHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	switch r.Method {
	case http.MethodGet:
		RenderFileTemplate(w, "register", db)
	case http.MethodPost:
		r.ParseForm()
		invalid := false
		if r.PostFormValue("username") == "" || r.PostFormValue("password") == "" {
			invalid = true
		} else {	
			for _, usr := range (*db).Users {
				if (*usr).Name == r.PostFormValue("username") {
					invalid = true
					break
				}
			}
		}
		if invalid {
			http.Redirect(w, r, "/registerfail", http.StatusTemporaryRedirect)
		} else {
			AddUser(
				r.PostFormValue("username"),
				r.PostFormValue("password"),
				r.PostFormValue("color"), 
				db)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
	}
}

//	webhandler for displaying the failed register page; redirects after 5 seconds
func RegisterFailHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	RenderFileTemplate(w, "regfail", db)
}

//	webhandler for followins a user
func FollowHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
	}
	authID, err := strconv.Atoi(session.Value)
	if err != nil {
		fmt.Println(err)
	}

	uID,_ := strconv.Atoi(r.URL.Path[len("/follow/"):])

	Follow(authID, uID, db)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//	webhandler for Deleting User, also deletes all of their Twoots
func DeleteHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
	}
	delID, err := strconv.Atoi(session.Value)
	if err != nil {
		fmt.Println(err)
	}

	DeleteUser(delID, db)

	LogoutHandler(w, r, db)
}


//	webhandler for Deleting Twoot, also resorts their twoots
func TDeleteHandler(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
	}
	authID, err := strconv.Atoi(session.Value)
	if err != nil {
		fmt.Println(err)
	}

	tID,_ := strconv.Atoi(r.URL.Path[len("/tdelete/"):])

	if (*(*(*db).Twoots[len((*db).Twoots) - tID]).Author).ID == authID {
		fmt.Printf("client owns TwootID: %d\n", tID)
		DeleteTwoot(tID, db)
		SortTwoots(&(*(*db).Users[authID]).Twoots)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//	function for sending out template for the client's timeline
func RenderTimeline(w http.ResponseWriter, r *http.Request, db *FakeDB) {
	session, err := r.Cookie("UserID")
	var inst Instance
	if err != nil {
		fmt.Println(err)
		inst = Instance{Client: nil, DB: db}
	} else {
		if session.Value == "" {
			inst = Instance{Client: nil, DB: db}
		} else {
			tempID, _ := strconv.Atoi(session.Value)
			tempUser := (*db).Users[tempID]
			timeline := FollowFilter((*tempUser).FollowList, db)
			inst = Instance{Client: tempUser, Timeline: timeline ,DB: db}
		}
	}

	head, err := template.ParseFiles("header.html")
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    content, err := template.ParseFiles("timeline.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    foot, err := template.ParseFiles("footer.html")
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = head.Execute(w, inst)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    err = content.Execute(w, inst)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    err = foot.Execute(w, inst)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

//	function for sending out specified template given its filename
func RenderFileTemplate(w http.ResponseWriter, tmpl string, db *FakeDB) {
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

	AddUser("Adam", "password", "#00afa0", &db)
	AddUser("Rick", "oo", "#859911", &db)
	AddUser("Ricardo", "pp", "#359890", &db)

	AddTwoot(0, "my last name is bouz", &db)
	AddTwoot(0, "what a nice day", &db)
	AddTwoot(0, "whats going on", &db)
	AddTwoot(1, "I like eggs", &db)
	AddTwoot(1, "did you see the game last night", &db)
	AddTwoot(1, "i know who im voting for in the election", &db)
	AddTwoot(2, "any movie recommendations", &db)
	AddTwoot(2, "the last episode of GOT was awesome", &db)
	AddTwoot(2, "check out this hilarious meme", &db)

	Follow(GetID("Adam", &db), GetID("Ricardo", &db), &db)

	http.HandleFunc("/", MakeDbHandler(BaseHandler, &db))
	http.HandleFunc("/login", MakeDbHandler(LoginHandler, &db))
	http.HandleFunc("/loginfail", MakeDbHandler(LoginFailHandler, &db))
	http.HandleFunc("/logout", MakeDbHandler(LogoutHandler, &db))
	http.HandleFunc("/post", MakeDbHandler(ComposeHandler, &db))
	http.HandleFunc("/register", MakeDbHandler(RegisterHandler, &db))
	http.HandleFunc("/registerfail", MakeDbHandler(RegisterFailHandler, &db))
	http.HandleFunc("/follow/", MakeDbHandler(FollowHandler, &db))
	http.HandleFunc("/delete", MakeDbHandler(DeleteHandler, &db))
	http.HandleFunc("/tdelete/", MakeDbHandler(TDeleteHandler, &db))

	fmt.Println("Initializing Server . . .")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
