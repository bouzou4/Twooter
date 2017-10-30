//	I apologize in advance for this mess
package main

import (
"fmt"
"os"
"bufio"
// "io/ioutil"
"net/http"
"html/template"
// "regexp"
"time"
"strings"
"strconv"
"crypto/sha256"
"encoding/hex"
)

//	regex to validate url request to be implemented soon(tm) 
//	var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

//	basic user struct with list of pointers to their created Twoots
type User struct {
	ID int
	Name string
	Pass string
	Color string
	FollowList []int
	// FollowedList []*User
	Twoots []int
}

type Twoot struct {
	ID int
	Author int
	Body string
	Created time.Time
}

//	no fs yet so the database is held in memory meaning memory violations are always a hair away
type FakeDB struct {
	Users []*User
	Twoots []*Twoot
}

type Instance struct {
	Client *User
	Timeline []*Twoot
	DB *FakeDB
}

func (db FakeDB) ParseUser(f *os.File) User {
	var lines []string
	fscan := bufio.NewScanner(f)
	for fscan.Scan() {
        lines = append(lines, fscan.Text())
    }

    follows := strings.Split(lines[4], " ")
    follows = follows[:len(follows) - 1]
    followed := strings.Split(lines[5], " ")
    followed = followed[:len(followed) - 1]

    p1, _ := strconv.Atoi(lines[0])

    var p5 = []int{}
    for _, x := range follows {
        y, _ := strconv.Atoi(x)
        p5 = append(p5, y)
    }

    var p6 = []int{}
    for _, x := range followed {
    	y, _ := strconv.Atoi(x)
        p6 = append(p6, y)
    }

    return User{ID: p1, Name: lines[1], Pass: lines[2], Color: lines[3], FollowList: p5, Twoots: p6}
}

func (db FakeDB) SaveUser(usr *User) string {
	data := strconv.Itoa(usr.ID) + "\n" + usr.Name + "\n" + usr.Pass + "\n" + usr.Color + "\n"

	for _, usr := range usr.FollowList {
		data += strconv.Itoa(usr) + " "
	}
	data += "\n"

	for _, twt := range usr.Twoots {
		data += strconv.Itoa(twt) + " "
	}
	data += "\n"

	return data
}

func (db FakeDB) WriteUsers() {
	for _, usr := range db.Users {
		tempPath := "Data/Users/" + strconv.Itoa(usr.ID) + ".txt"
		f, err := os.Create(tempPath)
		if err != nil {
			fmt.Println(err)
		}
		f.WriteString(db.SaveUser(usr))
		f.Close()
	}
}

func (db FakeDB) ParseTwoot(f *os.File) Twoot {
	var lines []string
	fscan := bufio.NewScanner(f)
	for fscan.Scan() {
        lines = append(lines, fscan.Text())
    }

    p1, _ := strconv.Atoi(lines[0])
    p2, _ := strconv.Atoi(lines[1])
    p4, _ := strconv.ParseInt(lines[3], 10, 64)

    return Twoot{ID: p1, Author: p2, Body: lines[3], Created: time.Unix(p4, 0)}
}

func (db FakeDB) SaveTwoot(twt *Twoot) string {
	return strconv.Itoa(twt.ID) + "\n" + strconv.Itoa(twt.Author) + "\n" + twt.Body + "\n" + strconv.FormatInt(twt.Created.Unix(), 10) + "\n"
}

func (db FakeDB) WriteTwoots() {
	for _, twt := range db.Twoots {
		tempPath := "Data/Twoots/" + strconv.Itoa(twt.ID) + ".txt"
		f, err := os.Create(tempPath)
		if err != nil {
			fmt.Println(err)
		}
		f.WriteString(db.SaveTwoot(twt))
		f.Close()
	}
}

func (db FakeDB) WriteDB() {
	var err error
	
	err = os.MkdirAll("Data/Users", 0755)
	if err != nil {
		fmt.Println(err)
	}
	db.WriteUsers()

	err = os.MkdirAll("Data/Twoots", 0755)
	if err != nil {
		fmt.Println(err)
	}
	db.WriteTwoots()

	index, err := os.Create("Data/Index.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer index.Close()

	index.WriteString("write 1\n")
	fmt.Fprintf(index, "hello my favorite number is %d\n", 4)
	index.Sync()
}

//	adds a user to the database while storing their password as a hash
//	returns UserID
func AddUser(name string, pass string, color string, db *FakeDB) int {
	h := sha256.New()
	h.Write([]byte(pass))
	bs := hex.EncodeToString(h.Sum(nil))

	tempID := len(db.Users)
	tempUser := &User{
		ID: tempID, 
		Name: name, 
		Pass: bs, 
		Color: color, 
		FollowList: []int{},
		Twoots: []int{},
	}
	db.Users = append(db.Users, tempUser)
	return tempID
}

//	creates and adds Twoot to the database
//	returns TwootID
func AddTwoot(author int, body string, db *FakeDB) int {
	tempID := len(db.Twoots)
	tempAuth := db.Users[author]
	tempTwoot := &Twoot{
		ID: tempID, 
		Author: author, 
		Body: body,
		Created: time.Now(),
	}
	
	tempTwoots := make([]*Twoot, len(db.Twoots) + 1)
	tempTwoots[0] = tempTwoot
	copy(tempTwoots[1:], db.Twoots)
	db.Twoots = tempTwoots

	tempaTwoots := make([]int, len((*tempAuth).Twoots) + 1)
	tempaTwoots[0] = tempTwoot.ID
	copy(tempaTwoots[1:], (*tempAuth).Twoots)
	(*tempAuth).Twoots = tempaTwoots

	return tempID
}

func Follow(user int, following int, db *FakeDB) {
	db.Users[user].FollowList = append(db.Users[user].FollowList, following)
	// db.Users[following].FollowedList = append(db.Users[following].FollowedList, db.Users[user])
}

//	filters out Twoots in database to find those asked for by follow list
//	returns list of pointers to them
func FollowFilter(follows []int, db *FakeDB) []*Twoot {
	timeline := []*Twoot{}
	for _, i := range db.Twoots {
		if GetID(follows, (*i).Author) != -1 {
			timeline = append(timeline, i)
		}
	}
	return timeline
}

//	resets all IDs in a list of Twoots to their proper order
func SortTwoots(list *[]*Twoot) {
	for i, x := range *list {
		x.ID = len(*list) - i
	}
	fmt.Println("sorted twoots")
}

//	Used to remove a Twoot from the DB given it's ID
func DeleteTwoot(dID int, db *FakeDB) {
	fmt.Printf("looking for Twoot: dID")
	for x := range db.Twoots {
		if (*db.Twoots[x]).ID == dID {
			fmt.Print("deleting twoot: ")
			fmt.Print((*db.Twoots[x]))
			fmt.Print("\n")
			copy(db.Twoots[x:], db.Twoots[x + 1:])
			db.Twoots[len(db.Twoots) - 1] = nil
			db.Twoots = db.Twoots[:len(db.Twoots) - 1]
			SortTwoots(&db.Twoots)
			break
		}
	}
}

//	Used to remove a User from the DB given their ID
func DeleteUser(delID int, db *FakeDB) {
	for x := range db.Users {
		if (*db.Users[x]).ID == delID {
			fmt.Printf("deleting user: %s\n", (*db.Users[x]).Name)
			for _, y := range (*db.Users[x]).Twoots {
				DeleteTwoot(y, db)
			}
			copy(db.Users[x:], db.Users[x + 1:])
			db.Users[len(db.Users) - 1] = nil
			db.Users = db.Users[:len(db.Users) - 1]
			break
		}
	}
}

//	itirates over array to find element
func GetID(sli []int, x int) int {
	for i, el := range sli {
		if el == x {
			return i
		}
	}
	return -1
}

//	function used to get userID from Username
//	returns -1 if not found
func GetUserID(username string, db *FakeDB) int {
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
	uID := GetUserID(username, db)
	h := sha256.New()
	h.Write([]byte(password))
	if hex.EncodeToString(h.Sum(nil)) == db.Users[uID].Pass {
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
			for _, usr := range db.Users {
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

	if db.Twoots[len(db.Twoots) - tID].Author == authID {
		fmt.Printf("client owns TwootID: %d\n", tID)
		DeleteTwoot(tID, db)
		//SortTwoots(&db.Users[authID].Twoots)
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
			tempUser := db.Users[tempID]
			timeline := FollowFilter(tempUser.FollowList, db)
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

//	function for sending out specified template given its fname
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

	Follow(GetUserID("Adam", &db), GetUserID("Ricardo", &db), &db)

	db.WriteDB()
	
	if 1==1 {
		j, err := os.Open("Data/Twoots/0.txt")
		if err != nil {
			fmt.Println(err)
		}
		defer j.Close()
		db.ParseTwoot(j)
	}

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
