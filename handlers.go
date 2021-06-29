package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

var taxedAmt = []float32{0.98, 0.67} //Fraction left after taxes
var jwtSignature string = os.Getenv("JWT_SIGNATURE")
var minEventsToRedeem int = 2

func SignUp(w http.ResponseWriter, r *http.Request) {

	user := &userDetails{}

	e := json.NewDecoder(r.Body).Decode(user)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.Rollno == "" {
		fmt.Fprint(w, "Roll Number should not be empty.")
		return
	}
	if user.Name == "" {
		fmt.Fprint(w, "Name should not be empty.")
		return
	}
	if len(user.Password) <= 7 {
		fmt.Fprint(w, "Password should be atleast 8 characters long.")
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
	er := json.NewDecoder(r.Body).Decode(user)

	if er != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var hashedPassword string
	var isAdmin bool

	sqlStmt := "SELECT password, isAdmin FROM User WHERE rollno = ?"
	er = Db.QueryRow(sqlStmt, user.Rollno).Scan(&hashedPassword, &isAdmin)

	if er != nil {
		fmt.Fprint(w, "Wrong username or password")
		return
	}

	er = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if er != nil {
		fmt.Fprint(w, "Wrong username or password")
		return
	}

	var batch string
	batch, er = getBatch(user.Rollno)

	if er != nil {
		fmt.Fprint(w, er.Error())
		return
	}

	token, err := GetToken(user.Rollno, isAdmin, batch)
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

	pl, err := isValidToken(r)

	if err != nil {
		fmt.Fprint(w, "Not Authorised.")
		return
	}

	res := map[string]interface{}{
		"message": fmt.Sprintf("Hello %s, This is a super secret information.", pl.Rollno),
	}

	json.NewEncoder(w).Encode(res)

}

func Reward(w http.ResponseWriter, r *http.Request) {

	pl, err := isValidToken(r)

	if err != nil || !pl.IsAdmin {
		fmt.Fprint(w, "Not Authorised.")
		return
	}

	user := &recipient{}
	json.NewDecoder(r.Body).Decode(user)

	if user.Coins <= 0 {
		fmt.Fprint(w, "Coins involved in a transaction must be positive!")
		return
	}

	var isAdmin bool

	isAdmin, err = AdminFlag(user.Rollno)

	if err != nil {
		logReward(user, false)
		fmt.Fprint(w, err.Error())
		return
	}

	if isAdmin {
		logReward(user, false)
		fmt.Fprint(w, "Invalid Request.")
		return
	}

	err = addCoins(user)

	if err != nil {
		logReward(user, false)
		fmt.Fprint(w, err.Error())
		return
	}

	logReward(user, true)
	fmt.Fprint(w, "Transaction Successful!")
}

func getCoins(w http.ResponseWriter, r *http.Request) {

	pl, er := isValidToken(r)

	if er != nil {
		fmt.Fprint(w, "Not Authorised.")
		return
	}

	rollNo := pl.Rollno

	if rollNo == "" {
		fmt.Fprint(w, "invalid roll number")
		return
	}

	coins, err := checkBalance(rollNo)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"coins": coins})
}

func Transfer(w http.ResponseWriter, r *http.Request) {

	// Validate token and get payload if the token is valid
	pl, err := isValidToken(r)

	if err != nil {
		fmt.Fprint(w, "Not Authorised.")
		return
	}
	// sender Roll number from the payload
	sender := pl.Rollno
	// check if the sender exists
	exist, er := Exists(sender)

	if er != nil {
		fmt.Fprint(w, er.Error())
		return
	}

	if !exist {
		fmt.Fprint(w, "Invalid request.")
		return
	}

	trf := &transfer{}
	json.NewDecoder(r.Body).Decode(trf)

	var toBatch, fromBatch string
	fromBatch, er = getBatch(sender)
	if er != nil {
		fmt.Fprint(w, er.Error())
		logTransfer(trf, false)
		return
	}
	toBatch, er = getBatch(trf.ToRollno)
	if er != nil {
		fmt.Fprint(w, er.Error())
		logTransfer(trf, false)
		return
	}
	i := 1
	if fromBatch == toBatch {
		i = 0
	}
	trf.TaxedAmt = taxedAmt[i]
	trf.FromRollno = sender

	if trf.Coins <= 0 {
		fmt.Fprint(w, "Coins involved in a transaction must be positive!")
		logTransfer(trf, false)
		return
	}

	var isAdmin bool

	isAdmin, err = AdminFlag(trf.ToRollno)

	if err != nil {
		fmt.Fprint(w, err.Error())
		logTransfer(trf, false)
		return
	}

	if isAdmin {
		fmt.Fprint(w, "Invalid Request.")
		logTransfer(trf, false)
		return
	}

	err = sendCoins(trf)

	if err != nil {
		fmt.Fprint(w, err.Error())
		logTransfer(trf, false)
		return
	}

	logTransfer(trf, true)
	fmt.Fprint(w, "Transaction Successful!")

}

func Redeem(w http.ResponseWriter, r *http.Request) {

	pl, err := isValidToken(r)

	if err != nil {
		fmt.Fprint(w, "Not Authorised.")
		return
	}

	red := &redeem{}
	json.NewDecoder(r.Body).Decode(red)

	red.Rollno = pl.Rollno

	var eventCount int
	err = Db.QueryRow("SELECT COUNT(*) FROM Transaction_Logs WHERE mode = ? AND secondaryUser = ?", "REWARD", red.Rollno).Scan(&eventCount)
	if err != nil {
		fmt.Fprint(w, "error: something went wrong")
		logRedeem(red, false)
		return
	}

	if eventCount < minEventsToRedeem {
		fmt.Print(w, "not eleigible to redeem yet")
		logRedeem(red, false)
		return
	}

	err = redeemCoins(red)

	if err != nil {
		fmt.Fprint(w, err.Error())
		logRedeem(red, false)
		return
	}

	logRedeem(red, true)
	fmt.Fprintf(w, "Transaction Successful!. Here are your %f coins", red.Coins)

}
