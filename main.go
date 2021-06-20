package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var err error
	Db, err = sql.Open("sqlite3", "students.db")
	if err != nil {
		panic(err)
	}

	statement, _ := Db.Prepare("CREATE TABLE IF NOT EXISTS User(rollno TEXT PRIMARY KEY, name TEXT NOT NULL, password TEXT NOT NULL, coins INT)") //Creating the Table
	statement.Exec()

	router := mux.NewRouter()

	router.HandleFunc("/signup", SignUp).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")
	router.HandleFunc("/secretpage", isAuthorised(Secret)).Methods("GET")
	router.HandleFunc("/transfer", Transfer).Methods("POST")
	router.HandleFunc("/reward", Reward).Methods("POST")
	router.HandleFunc("/view", getCoins).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))

}
