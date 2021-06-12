package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

func main() {
	var err error
	database, err = sql.Open("sqlite3", "students.db")
	if err != nil {
		panic(err)
	}

	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS User(rollno TEXT PRIMARY KEY, name TEXT NOT NULL, password TEXT NOT NULL)") //Creating the Table
	statement.Exec()

	http.HandleFunc("/signup", SignUp)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/secretpage", isAuthorised(Secret))

	log.Fatal(http.ListenAndServe(":8080", nil))

}
