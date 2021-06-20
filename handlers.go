package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

var jwtSignature string = os.Getenv("JWT_SIGNATURE")

func SignUp(w http.ResponseWriter, r *http.Request) {

	user := &userDetails{}

	e := json.NewDecoder(r.Body).Decode(user)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.Rollno == "" {
		fmt.Fprintf(w, "Roll Number should not be empty.")
		return
	}
	if user.Name == "" {
		fmt.Fprintf(w, "Name should not be empty.")
		return
	}
	if len(user.Password) <= 7 {
		fmt.Fprintf(w, "Password should be atleast 8 characters long.")
		return
	}

	passwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.Password = string(passwd)

	err = addUser(user)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, "User Successfully Added!")
}

func Login(w http.ResponseWriter, r *http.Request) {

	user := &userLogin{}
	e := json.NewDecoder(r.Body).Decode(user)

	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var hashedPassword string

	sqlStmt := "SELECT password FROM User WHERE rollno = ?"
	er := Db.QueryRow(sqlStmt, user.Rollno).Scan(&hashedPassword)

	if er != nil {
		fmt.Fprint(w, "Wrong username or password")
		return
	}

	er = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if er != nil {
		fmt.Fprint(w, "Wrong username or password")
		return
	}

	token, err := GetToken(user.Rollno)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"rollno": user.Rollno,
		"token":  token,
	}

	json.NewEncoder(w).Encode(res)

}

func Secret(w http.ResponseWriter, r *http.Request) {

	res := map[string]interface{}{
		"message": "This is a super secret information.",
	}

	json.NewEncoder(w).Encode(res)

}

func Reward(w http.ResponseWriter, r *http.Request) {

	user := &recipient{}
	json.NewDecoder(r.Body).Decode(user)

	if user.Coins <= 0 {
		fmt.Fprint(w, "Coins involved in a transaction must be positive!")
		return
	}

	err := addCoins(user)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, "Transaction Successful!")
}

func getCoins(w http.ResponseWriter, r *http.Request) {

	rollNo := &coinBal{}
	json.NewDecoder(r.Body).Decode(rollNo)

	coins, err := checkBalance(rollNo.Rollno)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"coins": coins})
}

func Transfer(w http.ResponseWriter, r *http.Request) {

	trf := &transaction{}
	json.NewDecoder(r.Body).Decode(trf)

	if trf.Coins <= 0 {
		fmt.Fprint(w, "Coins involved in a transaction must be positive!")
		return
	}

	err := sendCoins(trf)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, "Transaction Successful!")

}
