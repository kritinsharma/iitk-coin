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
	Db.SetMaxOpenConns(1)
	Db.Exec("CREATE TABLE IF NOT EXISTS User(rollno TEXT PRIMARY KEY, name TEXT NOT NULL, password TEXT NOT NULL, coins FLOAT, isAdmin BOOL)")
	Db.Exec("CREATE TABLE IF NOT EXISTS Transaction_Logs(transactionID INTEGER PRIMARY KEY AUTOINCREMENT, mode TEXT, primaryUser TEXT, deb_cred_primary FLOAT, secondaryUser TEXT, deb_cred_secondary FLOAT, madeAt DATE DEFAULT (DATETIME('now', 'localtime')), status BOOL)")

	router := mux.NewRouter()

	router.HandleFunc("/signup", SignUp).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")
	router.HandleFunc("/secretpage", Secret).Methods("GET")
	router.HandleFunc("/transfer", Transfer).Methods("POST")
	router.HandleFunc("/reward", Reward).Methods("POST")
	router.HandleFunc("/redeem", Redeem).Methods("POST")
	router.HandleFunc("/view", getCoins).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))

}
