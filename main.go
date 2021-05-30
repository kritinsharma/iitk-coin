package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type userDetails struct {
	rollno, name string
}

func addUser(db *sql.DB, usr *userDetails) {
	stmt, _ := db.Prepare("INSERT INTO User (rollno, name) VALUES (?, ?)") // Preparing the statement for adding the user
	_, err := stmt.Exec(usr.rollno, usr.name)

	if err != nil {
		panic(err)
	} else {
		fmt.Printf("User [%s: %s] succesfully added.", usr.rollno, usr.name)
	}
}

func main() {
	database, err := sql.Open("sqlite3", "students.db")
	if err != nil {
		panic(err)
	}

	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS User(rollno TEXT PRIMARY KEY, name TEXT NOT NULL)") //Creating the Table
	statement.Exec()

	//addUser(database, &userDetails{rollno: "Roll_no", name: "Name"})
}
