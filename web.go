//	I apologize in advance for this mess
package main

import (
"fmt"
"os"
"net"
"net/http"
"html/template"
// "regexp"
"time"
"strings"
"strconv"
"crypto/sha256"
"encoding/hex"
// "encoding/gob"
"bufio"
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
	Latest []*Twoot
	Users []*User
}

type AppServer struct {
	Connect net.Conn
	Scanr *bufio.Scanner
}

func (serv *AppServer) GetID(uID int) int {
	return 0
}

func (serv *AppServer) UserSearch(username string) int {
	return 0
}

func (serv *AppServer) GetUser(uID int) User {
	return User{}
}

func (serv *AppServer) GetNumUsers() int {
	return 0
}

func (serv *AppServer) GetUsers() []*User {
	return []*User{}
}

func (serv *AppServer) GetTwoot(uID int) Twoot {
	return Twoot{}
}

func (serv *AppServer) GetNumTwoots() int {
	return 0
}

func (serv *AppServer) GetTwoots() []*Twoot {
	return []*Twoot{}
}

func (serv *AppServer) Login(username string, password string) int {
	h := sha256.New()
	h.Write([]byte(password))
	hashed := hex.EncodeToString(h.Sum(nil))
	ret, _ := strconv.Atoi(serv.ServerRequest([]string{"Login", username, hashed}))
	return ret
}

func (serv *AppServer) AddTwoot(author int, body string) int {
	return 0
}

func (serv *AppServer) AddUser(name string, pass string, color string) int {
	return 0
}

func (serv *AppServer) DeleteTwoot(dID int) {

}

func (serv *AppServer) DeleteUser(delID int) {
	
}

func (serv *AppServer) Follow(user int, following int) {
	
}

func (serv *AppServer) Unfollow(user int, unfollowing int) {

}

func (serv *AppServer) FollowFilter(follows []int) []*Twoot {
	return []*Twoot{}
}

func (serv *AppServer) ServerRequest(args []string) string {
	// encoder := gob.NewEncoder(serv.Connect)
	// err := encoder.Encode(args)
	// if err != nil {
	// 	fmt.Printf("decode error: %s\n", err)
	// }

	fmt.Fprintln(serv.Connect, strings.Join(args[:]," ")) 

	scanner:= bufio.NewScanner(serv.Connect)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("server response to %s request: %s\n", args[0], line)
		return line
		break
	}		
	return ""
}

//	closure that returns a function that takes an http.ResponseWriter and http.Request and includes the FakeDB object
func MakeDbHandler(fn func(http.ResponseWriter, *http.Request, *AppServer), serv *AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, serv)
	}
}

//	webhandler for the homepage, if the user is logged in then they get their timeline 
//	otherwise they get the login page
func BaseHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
		RenderFileTemplate(w, "login")
	} else {

		tempID, err := strconv.Atoi(session.Value)
		if err != nil {
			fmt.Println(err)
			RenderFileTemplate(w, "login")
		} else if serv.GetID(tempID) == -1 {
			RenderFileTemplate(w, "login")
		} else {
			RenderTimeline(w, r, serv)
		}
	}
}

//	webhandler for login page performs login() and sets the cookie
func LoginHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	r.ParseForm()
	cookID := serv.Login(r.PostFormValue("username"), r.PostFormValue("password"))
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
func LoginFailHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	RenderFileTemplate(w, "loginfail")
}

//	webhandler for logout; essentially just clears cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	tok := http.Cookie {
		Name: "UserID",
		Value: "",
	}
	http.SetCookie(w, &tok)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//	webhandler for posting Twoots, will not post if text is longer than 100 chars
func ComposeHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
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
			serv.AddTwoot(author, r.PostFormValue("twoot"))
		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

//	webhandler for registering users and detecting if username is already in use
func RegisterHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	switch r.Method {
	case http.MethodGet:
		RenderFileTemplate(w, "register")
	case http.MethodPost:
		r.ParseForm()

		invalid := false
		if r.PostFormValue("username") == "" || r.PostFormValue("password") == "" {
			invalid = true
		} else if serv.UserSearch("username") != -1 {
			invalid = true
		}

		if invalid {
			http.Redirect(w, r, "/registerfail", http.StatusTemporaryRedirect)
		} else {
			serv.AddUser(
				r.PostFormValue("username"),
				r.PostFormValue("password"),
				r.PostFormValue("color"),
				)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
	}
}

//	webhandler for displaying the failed register page; redirects after 5 seconds
func RegisterFailHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	RenderFileTemplate(w, "regfail")
}

//	webhandler for followins a user
func FollowHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
	}
	authID, err := strconv.Atoi(session.Value)
	if err != nil {
		fmt.Println(err)
	}

	uID,_ := strconv.Atoi(r.URL.Path[len("/follow/"):])

	serv.Follow(authID, uID)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//	webhandler for followins a user
func UnfollowHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
	}
	authID, err := strconv.Atoi(session.Value)
	if err != nil {
		fmt.Println(err)
	}

	uID,_ := strconv.Atoi(r.URL.Path[len("/unfollow/"):])

	serv.Unfollow(authID, uID)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//	webhandler for Deleting User, also deletes all of their Twoots
func DeleteHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
	}
	delID, err := strconv.Atoi(session.Value)
	if err != nil {
		fmt.Println(err)
	}

	serv.DeleteUser(delID)

	LogoutHandler(w, r, serv)
}


//	webhandler for Deleting Twoot, also resorts their twoots
func TDeleteHandler(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	session, err := r.Cookie("UserID")
	if err != nil {
		fmt.Println(err)
	}
	authID, err := strconv.Atoi(session.Value)
	if err != nil {
		fmt.Println(err)
	}

	tID,_ := strconv.Atoi(r.URL.Path[len("/tdelete/"):])
	tempTwoot := serv.GetTwoot(tID)

	if tempTwoot.Author == authID {
		fmt.Printf("client %d is deleting TwootID: %d\n", authID, tID)
		serv.DeleteTwoot(tID)
		//SortTwoots(&db.Users[authID].Twoots)
	} else {
		fmt.Printf("client: %d attempted to delete invalid Twoot %s\n", authID, tempTwoot)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//	function for sending out template for the client's timeline
func RenderTimeline(w http.ResponseWriter, r *http.Request, serv *AppServer) {
	session, err := r.Cookie("UserID")
	var inst Instance
	if err != nil {
		fmt.Println(err)
	} else {
		if session.Value != "" {
			tempID, _ := strconv.Atoi(session.Value)
			tempUser := serv.GetUser(tempID)
			timeline := serv.FollowFilter(tempUser.FollowList)
			latest := serv.GetTwoots()
			users := serv.GetUsers()
			inst = Instance{Client: &tempUser, Timeline: timeline, Latest: latest, Users: users}
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
func RenderFileTemplate(w http.ResponseWriter, tmpl string) {
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
	err = head.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = content.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = foot.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	service := "localhost:8083"
	port := ":8080"

	conn, err := net.Dial("tcp", service)
	if err != nil {
		fmt.Fprint(os.Stderr, "could not connect", err.Error())
	}

	serv := AppServer{Connect: conn}

	http.HandleFunc("/", MakeDbHandler(BaseHandler, &serv))
	http.HandleFunc("/login", MakeDbHandler(LoginHandler, &serv))
	http.HandleFunc("/loginfail", MakeDbHandler(LoginFailHandler, &serv))
	http.HandleFunc("/logout", MakeDbHandler(LogoutHandler, &serv))
	http.HandleFunc("/post", MakeDbHandler(ComposeHandler, &serv))
	http.HandleFunc("/register", MakeDbHandler(RegisterHandler, &serv))
	http.HandleFunc("/registerfail", MakeDbHandler(RegisterFailHandler, &serv))
	http.HandleFunc("/follow/", MakeDbHandler(FollowHandler, &serv))
	http.HandleFunc("/unfollow/", MakeDbHandler(UnfollowHandler, &serv))
	http.HandleFunc("/delete", MakeDbHandler(DeleteHandler, &serv))
	http.HandleFunc("/tdelete/", MakeDbHandler(TDeleteHandler, &serv))

	fmt.Println("Initializing Server . . .")
	// fmt.Fprintf(serv.Connect, "Initializing Web Server at %s\n", port)
	fmt.Println(http.ListenAndServe(port, nil))
}
