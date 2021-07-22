package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	godotenv "github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var fromEmail string
var emailPasswd string
var jwtSignature string

func main() {
	er := godotenv.Load()
	if er != nil {
		log.Fatal("Error loading .env file")
	}

	fromEmail = os.Getenv("FROM_EMAIL")
	emailPasswd = os.Getenv("EMAIL_PASSWORD")
	jwtSignature = os.Getenv("JWT_SIGNATURE")

	var err error
	Db, err = sql.Open("sqlite3", "students.db")
	if err != nil {
		panic(err)
	}

	Db.SetMaxOpenConns(1)
	Db.Exec("CREATE TABLE IF NOT EXISTS User(rollno TEXT PRIMARY KEY, name TEXT NOT NULL, password TEXT NOT NULL, email TEXT, coins FLOAT, isAdmin BOOL)")
	Db.Exec("CREATE TABLE IF NOT EXISTS Transaction_Logs(transactionID INTEGER PRIMARY KEY AUTOINCREMENT, mode TEXT, primaryUser TEXT, deb_cred_primary FLOAT, secondaryUser TEXT, deb_cred_secondary FLOAT, madeAt DATE DEFAULT (DATETIME('now', 'localtime')))")
	Db.Exec("CREATE TABLE IF NOT EXISTS RedeemRequests(requestID INTEGER PRIMARY KEY AUTOINCREMENT, itemName TEXT, coins FLOAT, madeBy TEXT, madeAt DATE DEFAULT (DATETIME('now', 'localtime')), status VARCHAR(1) DEFAULT 'p')")
	router := mux.NewRouter()

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	router.HandleFunc("/signup", SignUp).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")
	router.HandleFunc("/secretpage", Secret).Methods("GET")
	router.HandleFunc("/transfer", Transfer).Methods("POST")
	router.HandleFunc("/reward", Reward).Methods("POST")
	router.HandleFunc("/redeem", Redeem).Methods("POST")
	router.HandleFunc("/view", getCoins).Methods("GET")
	router.HandleFunc("/pending", getPendingRequests).Methods("GET")
	router.HandleFunc("/accept-reject", AcceptRejectRedeemRequest).Methods("POST")
	router.HandleFunc("/getOTP", SendOTP).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))

}
