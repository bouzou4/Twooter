package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Page struct {
    Title string
    Body  []byte
}

//This is a method named save that takes as its receiver p, a pointer to Page. It takes no parameters, and returns a value of type error
func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

// constructs the file name from the title parameter, reads the file's contents into a new variable body, and returns a pointer to a Page literal constructed with the proper title and body values.
func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
  	title := r.URL.Path[len("/view/"):]
  	p, _ := loadPage(title)
  	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
  }

func main() {
	p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
    p1.save()
    p2, _ := loadPage("TestPage")
    fmt.Println(string(p2.Body))

    http.HandleFunc("/view/", viewHandler)
    fmt.Println(http.ListenAndServe(":8080", nil))
}
