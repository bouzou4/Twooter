//	I apologize in advance for this mess
package main

import (
"fmt"
"os"
"bufio"
// "io/ioutil"
"net"
// "regexp"
"time"
"strings"
"strconv"
"crypto/sha256"
"encoding/hex"
// "encoding/gob"
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
	DB *FakeDB
}

func readLines(path string) []string {
	f, err := os.Open(path);
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	var lines []string
	fscan := bufio.NewScanner(f)
	for fscan.Scan() {
		lines = append(lines, fscan.Text())
	}
	return lines
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

func UserSearch(username string, db *FakeDB) int {
	for _, el := range db.Users {
		if el.Name == username {
			return el.ID
		}
	}
	return -1
}

func ReverseTwoots(inp []*Twoot) *[]*Twoot {
	ret := []*Twoot{}
	for i := len(inp) - 1; i >= 0; i-- {
		ret = append(ret, inp[i])
	}
	return &ret
}

func (db *FakeDB) LoadDB() {
	f, err := os.Open("Data/Index.txt"); 
	if err == nil {
		f.Close()
		lines := readLines("Data/Index.txt")
		numUsers, err := strconv.Atoi(lines[1])
		if err != nil {
			fmt.Println(err)
		} 
		numTwoots, err := strconv.Atoi(lines[4])
		if err != nil {
			fmt.Println(err)
		}
		for i := 0; i < numUsers; i++ {
			db.Users = append(db.Users, ParseUser(i))
		}
		for i := 0; i < numTwoots; i++ {
			db.Twoots = append(db.Twoots, ParseTwoot(i))
		}
	} else {
		fmt.Println(err)
	}
}

func ParseUser(i int) *User {
	lines := readLines(fmt.Sprintf("Data/Users/%d.txt", i))

	p1, _ := strconv.Atoi(lines[0])

	var p5 = []int{}
	var p6 = []int{}

	if lines[4] != ""	{
		follows := strings.Split(lines[4], " ")
		follows = follows[:len(follows) - 1]
		
		for _, x := range follows {
			y, _ := strconv.Atoi(x)
			p5 = append(p5, y)
		}
	}

	if lines[5] != ""	{
		twoots := strings.Split(lines[5], " ")
		twoots = twoots[:len(twoots) - 1]
		
		for _, x := range twoots {
			y, _ := strconv.Atoi(x)
			p6 = append(p6, y)
		}
	}

	return &User{ID: p1, Name: lines[1], Pass: lines[2], Color: lines[3], FollowList: p5, Twoots: p6}
}

func (db *FakeDB) SaveUser(usr *User) string {
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

func (db *FakeDB) WriteUsers() {
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

func ParseTwoot(i int) *Twoot {
	lines := readLines(fmt.Sprintf("Data/Twoots/%d.txt", i))

	p1, _ := strconv.Atoi(lines[0])
	p2, _ := strconv.Atoi(lines[1])
	p4, _ := strconv.ParseInt(lines[3], 10, 64)

	return &Twoot{ID: p1, Author: p2, Body: lines[2], Created: time.Unix(p4, 0)}
}

func (db *FakeDB) SaveTwoot(twt *Twoot) string {
	return strconv.Itoa(twt.ID) + "\n" + strconv.Itoa(twt.Author) + "\n" + twt.Body + "\n" + strconv.FormatInt(twt.Created.Unix(), 10) + "\n"
}

func (db *FakeDB) WriteTwoots() {
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

func (db *FakeDB) WriteDB() {
	var err error

	fmt.Println("updating server . . .")
	
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

	fmt.Fprintf(index, "Users\n%d\n\nTwoots\n%d\n", len(db.Users), len(db.Twoots))
	index.Sync()

	fmt.Println("server updated")
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

	db.WriteDB()
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
	
	db.Twoots = append(db.Twoots, tempTwoot)
	tempAuth.Twoots = append(tempAuth.Twoots, tempTwoot.ID)

	db.WriteDB()
	return tempID
}

func Follow(user int, following int, db *FakeDB) {
	if GetID(db.Users[user].FollowList, following) == -1 {
		db.Users[user].FollowList = append(db.Users[user].FollowList, following)
		db.WriteDB()
	}
	// db.Users[following].FollowedList = append(db.Users[following].FollowedList, db.Users[user])
}

func Unfollow(user int, unfollowing int, db *FakeDB) {
	ind := GetID(db.Users[user].FollowList, unfollowing)
	if ind != -1 {
		copy(db.Users[user].FollowList[ind:], db.Users[user].FollowList[ind+1:])
		db.Users[user].FollowList[len(db.Users[user].FollowList)-1] = -1 // or the zero value of T
		db.Users[user].FollowList = db.Users[user].FollowList[:len(db.Users[user].FollowList)-1]

		db.WriteDB()
	}
	// db.Users[following].FollowedList = append(db.Users[following].FollowedList, db.Users[user])
}


//	filters out Twoots in database to find those asked for by follow list
//	returns list of pointers to them
func FollowFilter(follows []int, db *FakeDB) []*Twoot {
	timeline := []*Twoot{}
	for _, x := range *ReverseTwoots(db.Twoots) {
		if GetID(follows, x.Author) != -1 {
			timeline = append(timeline, x)
		}
	}
	return timeline
}

//	resets all IDs in a list of Twoots to their proper order
func SortTwoots(list *[]*Twoot, db *FakeDB) {
	for i, x := range *list {
		x.ID = i
	}
	db.WriteDB()
	fmt.Println("sorted twoots")
}

//	Used to remove a Twoot from the DB given it's ID
func DeleteTwoot(dID int, db *FakeDB) {
	fmt.Printf("looking for Twoot: %d\n", dID)
	if db.Twoots[dID].ID == dID {
		fmt.Print("deleting twoot: ")
		fmt.Println(db.Twoots[dID])

		copy(db.Twoots[dID:], db.Twoots[dID + 1:])
		db.Twoots[len(db.Twoots) - 1] = nil
		db.Twoots = db.Twoots[:len(db.Twoots) - 1]

		err := os.Remove(fmt.Sprintf("Data/Twoots/%d.txt", len(db.Twoots)))
		if err != nil {
			fmt.Println(err)
		}
	}
	SortTwoots(&db.Twoots, db)
	db.WriteDB()
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
	db.WriteDB()
}

//	function used to get userID from Username
//	returns -1 if not found
func GetUserID(username string, db *FakeDB) int {
	for _, usr := range db.Users {
		if usr.Name == username {
			return usr.ID
		}
	}
	fmt.Printf("couldn't find user: %s in db of %s\n", username, db.Users)
	return -1
}

//	function used to check if username and password hash match up
//	returns -1 if credentials are invalid
func login(username string, hashed string, db *FakeDB) int {
	uID := GetUserID(username, db)
	if uID >= 0 && uID < len(db.Users) {
		if hashed == db.Users[uID].Pass {
			return uID
		}
		fmt.Printf("attempted login with incorrect password\n")
		return -1
	}
	fmt.Printf("attempted login with incorrect id: %d\n", uID)
	return -1
}

func handleConnection(Connect net.Conn, db *FakeDB) {
	// fmt.Println("begin handling");
	// dec := gob.NewDecoder(Connect)
	// var p [4]string
	// err := dec.Decode(&p)
	// if err != nil {
	// 	fmt.Printf("decode error: %s\n", err)
	// }
	// fmt.Println(p);
	scanner:= bufio.NewScanner(Connect)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("received request: %s\n", line)
		args :=  strings.Split(line, " ")
		
		switch args[0] {
		case "Login":
			fmt.Fprintln(Connect, strconv.Itoa(login(args[1], args[2], db)))
		default:
			fmt.Println("invalid request made: %s\n", args[0])
		}
	}

}

func main() {
	db := FakeDB{Users: []*User{}, Twoots: []*Twoot{}}
	db.LoadDB()

	req, err := net.Listen("tcp", ":8083")
	
	if err != nil {
		fmt.Println(err)
	}
	defer req.Close()

	for {
		conn, err := req.Accept()
		if err != nil {
			fmt.Fprint(os.Stderr, "Failed to accept")
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "accept successful")

		handleConnection(conn, &db)

		defer fmt.Fprintln(os.Stderr, "connection gone!")
		defer conn.Close()
	}
}
