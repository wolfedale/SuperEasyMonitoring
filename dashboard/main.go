package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

//
var database *sql.DB

//
const dbpath = "/Users/pawel/Private/Code/golang/SuperEasyMonitoring/conf/foo.db"

// Status type struct for Status
type Status struct {
	ID               int
	Host             string
	Checkname        string
	Status           string
	InsertedDatetime string
}

// customErr is for returning error so we don't need to wring it every time
// when we have any err variable, we can simply call this one.
func customErr(text string, e error) {
	if e != nil {
		fmt.Println(text, e)
		os.Exit(1)
	}
}

// ReadItem bla bla bla
func ReadItem(db *sql.DB) []Status {
	sqlReadall := `
		SELECT ID, Hostname, Checkname, Status, InsertedDatetime FROM Monitoring ORDER BY ID
		`
	rows, err := db.Query(sqlReadall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []Status
	for rows.Next() {
		item := Status{}
		err2 := rows.Scan(&item.ID, &item.Host, &item.Checkname, &item.Status, &item.InsertedDatetime)
		if err2 != nil {
			panic(err2)
		}
		//fmt.Println(item)
		result = append(result, item)
	}
	return result
}

func dashboard(db *sql.DB) {
	tmpl := template.Must(template.ParseFiles("html/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := ReadItem(db)
		tmpl.Execute(w, data)
	})

	// get CSS
	http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("html"))))
	http.ListenAndServe(":3000", nil)
}

// InitDB initialize database
func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	database = db
	return database
}

func main() {
	// database
	db := InitDB(dbpath)
	defer db.Close()

	// dashboard
	dashboard(db)
}
