//	I apologize in advance for this mess
package main

import (
	"bufio"
	"fmt"
	"os"
	// "io/ioutil"
	"net"
	"sync"
	// "regexp"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
	// "encoding/gob"
)

//	regex to validate url request to be implemented soon(tm)
//	var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

//	basic user struct with list of the IDs of their Twoots
type User struct {
	ID         int
	Name       string
	Pass       string
	Color      string
	FollowList []int
	Twoots     []int
	mut        sync.RWMutex
}

type Twoot struct {
	ID      int
	Author  int
	Body    string
	Created time.Time
	mut     sync.RWMutex
}

//	memory representation of database
type MemDB struct {
	Users  []*User
	Twoots []*Twoot
	umut   sync.RWMutex
	tmut   sync.RWMutex
}

//	instance for timeline template containing relevant information
type Instance struct {
	Client   *User
	Timeline []*Twoot
	Latest   []*Twoot
	DB       *MemDB
}

//	takes in a path to a text file and returns a string array of each line
func readLines(path string) []string {
	f, err := os.Open(path)
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

//	itirates over array of ints to find element
// 	if found returns index otherwise returns -1
func GetID(sli []int, x int) int {
	for i, el := range sli {
		if el == x {
			return i
		}
	}
	return -1
}

//	searches for ID of user given their username
// 	if found returns their ID otherwise returns -1
func UserSearch(username string, db *MemDB) int {
	db.umut.RLock()
	for _, el := range db.Users {
		if el.Name == username {
			return el.ID
		}
	}
	db.umut.RUnlock()
	return -1
}

//	reverses slice of Twoot pointers
func ReverseTwoots(inp []*Twoot) *[]*Twoot {
	ret := []*Twoot{}
	for i := len(inp) - 1; i >= 0; i-- {
		ret = append(ret, inp[i])
	}
	return &ret
}

// 	creates MemDB from file system
func (db *MemDB) LoadDB() {
	f, err := os.Open("Data/Index.txt")
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

// 	creates User from passed index of User's file in file system
func ParseUser(i int) *User {
	lines := readLines(fmt.Sprintf("Data/Users/%d.txt", i))

	p1, _ := strconv.Atoi(lines[0])

	var p5 = []int{}
	var p6 = []int{}

	if lines[4] != "" {
		follows := strings.Split(lines[4], " ")
		follows = follows[:len(follows)-1]

		for _, x := range follows {
			y, _ := strconv.Atoi(x)
			p5 = append(p5, y)
		}
	}

	if lines[5] != "" {
		twoots := strings.Split(lines[5], " ")
		twoots = twoots[:len(twoots)-1]

		for _, x := range twoots {
			y, _ := strconv.Atoi(x)
			p6 = append(p6, y)
		}
	}

	return &User{ID: p1, Name: lines[1], Pass: lines[2], Color: lines[3], FollowList: p5, Twoots: p6}
}

// 	encodes User object, given its ID, to a string with its field joined by specified connector
func (db *MemDB) SaveUser(uID int, connector string) string {
	usr := db.Users[uID]
	data := strconv.Itoa(usr.ID) + connector + usr.Name + connector + usr.Pass + connector + usr.Color + connector

	for _, usr := range usr.FollowList {
		data += strconv.Itoa(usr) + " "
	}
	data += connector

	for _, twt := range usr.Twoots {
		data += strconv.Itoa(twt) + " "
	}
	data += connector

	return data
}

// returns all Users in database as an encoded string
func (db *MemDB) SendUsers() string {
	ret := ""
	db.umut.RLock()
	for _, usr := range db.Users {
		ret += db.SaveUser(usr.ID, "|")
		ret += "[|]"
	}
	db.umut.RUnlock()
	return ret
}

// 	logs all users to the filesystem
func (db *MemDB) WriteUsers() {
	for _, usr := range db.Users {
		tempPath := "Data/Users/" + strconv.Itoa(usr.ID) + ".txt"
		f, err := os.Create(tempPath)
		if err != nil {
			fmt.Println(err)
		}
		f.WriteString(db.SaveUser(usr.ID, "\n"))
		f.Close()
	}
}

// 	creates Twoot from passed index of Twoot's file in file system
func ParseTwoot(i int) *Twoot {
	lines := readLines(fmt.Sprintf("Data/Twoots/%d.txt", i))

	p1, _ := strconv.Atoi(lines[0])
	p2, _ := strconv.Atoi(lines[1])
	p4, _ := strconv.ParseInt(lines[3], 10, 64)

	return &Twoot{ID: p1, Author: p2, Body: lines[2], Created: time.Unix(p4, 0)}
}

// 	encodes Twoot object, given its ID, to a string with its field joined by specified connector
func (db *MemDB) SaveTwoot(tID int, connector string) string {
	if tID >= 0 && tID < len(db.Twoots) {
		twt := db.Twoots[tID]
		return strconv.Itoa(twt.ID) + connector + strconv.Itoa(twt.Author) + connector + twt.Body + connector + strconv.FormatInt(twt.Created.Unix(), 10) + connector
	}
	return ""
}

// returns all Twoots in database as an encoded string
func (db *MemDB) SendTwoots(reversed bool) string {
	ret := ""
	if !reversed {
		db.tmut.RLock()
		for _, twt := range db.Twoots {
			ret += db.SaveTwoot(twt.ID, "|")
			ret += "[|]"
		}
		db.tmut.RUnlock()
	} else {
		db.tmut.RLock()
		for _, twt := range *ReverseTwoots(db.Twoots) {
			ret += db.SaveTwoot(twt.ID, "|")
			ret += "[|]"
		}
		db.tmut.RUnlock()
	}

	return ret
}

// 	logs all twoots to the filesystem
func (db *MemDB) WriteTwoots() {
	for _, twt := range db.Twoots {
		tempPath := "Data/Twoots/" + strconv.Itoa(twt.ID) + ".txt"
		f, err := os.Create(tempPath)
		if err != nil {
			fmt.Println(err)
		}
		f.WriteString(db.SaveTwoot(twt.ID, "\n"))
		f.Close()
	}
}

// 	writes all changes to database
func (db *MemDB) WriteDB() {
	var err error

	fmt.Println("updating server . . .")

	err = os.MkdirAll("Data/Users", 0755)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("getting user lock . . .")
	db.umut.RLock()
	db.WriteUsers()
	db.umut.RUnlock()

	err = os.MkdirAll("Data/Twoots", 0755)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("getting twoot lock . . .")
	db.tmut.RLock()
	db.WriteTwoots()
	db.tmut.RUnlock()

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
func AddUser(name string, pass string, color string, db *MemDB) int {
	h := sha256.New()
	h.Write([]byte(pass))
	bs := hex.EncodeToString(h.Sum(nil))

	db.umut.Lock()
	tempID := len(db.Users)
	tempUser := &User{
		ID:         tempID,
		Name:       name,
		Pass:       bs,
		Color:      color,
		FollowList: []int{},
		Twoots:     []int{},
	}
	db.Users = append(db.Users, tempUser)
	db.umut.Unlock()

	db.WriteDB()
	return tempID
}

//	creates and adds Twoot to the database
//	returns TwootID
func AddTwoot(author int, body string, db *MemDB) int {
	db.tmut.Lock()
	fmt.Println("1")
	db.umut.RLock()
	fmt.Println("2")
	tempID := len(db.Twoots)
	tempAuth := db.Users[author]
	db.umut.RUnlock()
	tempAuth.mut.Lock()
	fmt.Println("3")
	tempTwoot := &Twoot{
		ID:      tempID,
		Author:  author,
		Body:    body,
		Created: time.Now(),
	}

	tempAuth.Twoots = append(tempAuth.Twoots, tempTwoot.ID)
	tempAuth.mut.Unlock()
	db.Twoots = append(db.Twoots, tempTwoot)
	db.tmut.Unlock()

	db.WriteDB()
	return tempID
}

func Follow(user int, following int, db *MemDB) {
	if GetID(db.Users[user].FollowList, following) == -1 {
		db.Users[user].FollowList = append(db.Users[user].FollowList, following)
		db.WriteDB()
	}
	// db.Users[following].FollowedList = append(db.Users[following].FollowedList, db.Users[user])
}

func Unfollow(user int, unfollowing int, db *MemDB) {
	ind := GetID(db.Users[user].FollowList, unfollowing)
	if ind != -1 {
		copy(db.Users[user].FollowList[ind:], db.Users[user].FollowList[ind+1:])
		db.Users[user].FollowList[len(db.Users[user].FollowList)-1] = -1 // or the zero value of T
		db.Users[user].FollowList = db.Users[user].FollowList[:len(db.Users[user].FollowList)-1]

		db.WriteDB()
	}
	// db.Users[following].FollowedList = append(db.Users[following].FollowedList, db.Users[user])
}

//	resets all IDs in a list of Twoots to their proper order
func SortTwoots(list *[]*Twoot, db *MemDB) {
	for i, x := range *list {
		x.ID = i
	}
	fmt.Println("sorted twoots")
}

//	Used to remove a Twoot from the DB given it's ID
func DeleteTwoot(dID int, db *MemDB) {
	fmt.Printf("looking for Twoot: %d\n", dID)
	db.tmut.Lock()
	if dID >= 0 && dID < len(db.Twoots) {
		if db.Twoots[dID].ID == dID {
			fmt.Printf("deleting twoot: \n%s\n", db.Twoots[dID])

			copy(db.Twoots[dID:], db.Twoots[dID+1:])
			db.Twoots[len(db.Twoots)-1] = nil
			db.Twoots = db.Twoots[:len(db.Twoots)-1]

			err := os.Remove(fmt.Sprintf("Data/Twoots/%d.txt", len(db.Twoots)))
			if err != nil {
				fmt.Println(err)
			}
		}
		SortTwoots(&db.Twoots, db)
	}
	db.tmut.Unlock()
	db.WriteDB()
}

//	Used to remove a User from the DB given their ID
func DeleteUser(delID int, db *MemDB) {
	if delID >= 0 && delID < len(db.Users) {
		db.umut.Lock()
		for i, x := range db.Users {
			if x.ID == delID {
				x.mut.Lock()
				fmt.Printf("deleting user: %s\n", x.Name)
				for _, y := range x.Twoots {
					DeleteTwoot(y, db)
				}
				x.mut.Unlock()

				copy(db.Users[i:], db.Users[i+1:])
				db.Users[len(db.Users)-1] = nil
				db.Users = db.Users[:len(db.Users)-1]

				err := os.Remove(fmt.Sprintf("Data/Users/%d.txt", len(db.Users)))
				if err != nil {
					fmt.Println(err)
				}
			} else {
				if GetID(x.FollowList, delID) != -1 {
					Unfollow(x.ID, delID, db)
				}
			}
			db.umut.Unlock()
		}
	}
}

//	function used to get userID from Username
//	returns -1 if not found
func GetUserID(username string, db *MemDB) int {
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
func login(username string, hashed string, db *MemDB) int {
	db.umut.RLock()
	uID := GetUserID(username, db)
	defer db.umut.RUnlock()
	if uID >= 0 && uID < len(db.Users) {
		if hashed == db.Users[uID].Pass {
			fmt.Printf("User %s logged in\n", username)
			return uID
		}
		fmt.Printf("attempted login with incorrect password\n")
		return -1
	}
	fmt.Printf("attempted login with incorrect id: %d\n", uID)
	return -1
}

func handleConnection(Connect net.Conn, db *MemDB) {
	// fmt.Println("begin handling");
	// dec := gob.NewDecoder(Connect)
	// var p [4]string
	// err := dec.Decode(&p)
	// if err != nil {
	// 	fmt.Printf("decode error: %s\n", err)
	// }
	// fmt.Println(p);
	scanner := bufio.NewScanner(Connect)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("received request: %s\n", strings.Join(strings.Split(line, "[}{]"), ", "))
		args := strings.Split(line, "[}{]")

		switch args[0] {
		case "Login":
			go fmt.Fprintln(Connect, strconv.Itoa(login(args[1], args[2], db)))
		case "GetID":
			if args[1] == "Users" {
				id, _ := strconv.Atoi(args[2])
				db.umut.RLock()
				fmt.Println("got Users Lock")
				if !(id >= 0 && id < len(db.Users)) {
					fmt.Fprintln(Connect, strconv.Itoa(id))
				} else {
					fmt.Fprintln(Connect, strconv.Itoa(-1))
				}
				db.umut.RUnlock()
				fmt.Println("gave up Users Lock")
			} else if args[1] == "Twoots" {
				id, _ := strconv.Atoi(args[2])
				db.tmut.RLock()
				fmt.Println("got Twoots Lock")
				if !(id >= 0 && id < len(db.Twoots)) {
					fmt.Fprintln(Connect, strconv.Itoa(id))
				} else {
					fmt.Fprintln(Connect, strconv.Itoa(-1))
				}
				db.tmut.RUnlock()
				fmt.Println("goave upTwoots Lock")
			} else {
				fmt.Fprintln(Connect, strconv.Itoa(-1))
			}

		case "UserSearch":
			go fmt.Fprintln(Connect, strconv.Itoa(UserSearch(args[1], db)))
		case "GetUser":
			ind, _ := strconv.Atoi(args[1])
			db.umut.RLock()
			fmt.Fprintln(Connect, db.SaveUser(ind, "|"))
			db.umut.RUnlock()
		case "GetNumUsers":
			go fmt.Fprintln(Connect, len(db.Users))
		case "GetUsers":
			go fmt.Fprintln(Connect, db.SendUsers())
		case "GetTwoot":
			ind, _ := strconv.Atoi(args[1])
			db.tmut.RLock()
			fmt.Fprintln(Connect, db.SaveTwoot(ind, "|"))
			db.tmut.RUnlock()
		case "GetNumTwoots":
			go fmt.Fprintln(Connect, len(db.Twoots))
		case "GetTwoots":
			rev, _ := strconv.ParseBool(args[1])
			go fmt.Fprintln(Connect, db.SendTwoots(rev))
		case "AddTwoot":
			id, _ := strconv.Atoi(args[1])
			go fmt.Fprintln(Connect, strconv.Itoa(AddTwoot(id, args[2], db)))
		case "AddUser":
			go fmt.Fprintln(Connect, AddUser(args[1], args[2], args[3], db))
		case "DeleteTwoot":
			ind, _ := strconv.Atoi(args[1])
			go DeleteTwoot(ind, db)
			fmt.Fprintln(Connect, "Done")
		case "DeleteUser":
			ind, _ := strconv.Atoi(args[1])
			DeleteUser(ind, db)
			fmt.Fprintln(Connect, "Done")
		case "Follow":
			ind1, _ := strconv.Atoi(args[1])
			ind2, _ := strconv.Atoi(args[2])
			db.umut.RLock()
			go Follow(ind1, ind2, db)
			db.umut.RUnlock()
			fmt.Fprintln(Connect, "Done")
		case "Unfollow":
			ind1, _ := strconv.Atoi(args[1])
			ind2, _ := strconv.Atoi(args[2])
			db.umut.RLock()
			go Unfollow(ind1, ind2, db)
			db.umut.RUnlock()
			fmt.Fprintln(Connect, "Done")

		default:
			fmt.Println("invalid request made: %s\n", args[0])
		}
	}

}

func main() {
	db := MemDB{Users: []*User{}, Twoots: []*Twoot{}}
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
