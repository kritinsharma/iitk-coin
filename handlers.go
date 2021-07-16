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
	if len(user.Password) <= 7 || len(user.Password) > 72 {
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

	exists, _ := Exists(pl.Rollno)

	if err != nil || !exists {
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

	exists, _ := Exists(pl.Rollno)

	if err != nil || !exists || !pl.IsAdmin {
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
		fmt.Fprint(w, err.Error())
		return
	}

	if isAdmin {
		fmt.Fprint(w, "Invalid Request.")
		return
	}

	err = addCoins(user)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, "Transaction Successful!")
}

func getCoins(w http.ResponseWriter, r *http.Request) {

	pl, err := isValidToken(r)

	exists, _ := Exists(pl.Rollno)

	if err != nil || !exists {
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

	exists, _ := Exists(pl.Rollno)

	if err != nil || !exists {
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
		return
	}
	toBatch, er = getBatch(trf.ToRollno)
	if er != nil {
		fmt.Fprint(w, er.Error())
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
		return
	}

	var isAdmin bool

	isAdmin, err = AdminFlag(trf.ToRollno)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	if isAdmin {
		fmt.Fprint(w, "Invalid Request.")
		return
	}

	err = sendCoins(trf)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, "Transaction Successful!")

}

func Redeem(w http.ResponseWriter, r *http.Request) {

	pl, err := isValidToken(r)

	exists, _ := Exists(pl.Rollno)

	if err != nil || !exists {
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
		return
	}

	if eventCount < minEventsToRedeem {
		fmt.Fprint(w, "not eleigible to redeem yet")
		return
	}

	err = createRedeemRequest(red)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, "Your redeem request status: pending. It will get updated when the request is approved/rejected")

}

func getPendingRequests(w http.ResponseWriter, r *http.Request) {

	pl, err := isValidToken(r)

	exists, _ := Exists(pl.Rollno)

	if err != nil || !exists || !pl.IsAdmin {
		fmt.Fprint(w, "Not Authorised.")
		return
	}

	var pendingReqs []redeemRequest

	rows, er := Db.Query("SELECT requestID, itemName, coins, madeBy, madeAt FROM RedeemRequests WHERE status = 'p'")
	if er != nil {
		fmt.Println(er)
		fmt.Fprint(w, "internal error")
		return
	}
	defer rows.Close()

	var pendingReq redeemRequest
	pendingReq.Status = "p"

	for rows.Next() {
		er = rows.Scan(&pendingReq.RequestID, &pendingReq.ItemName, &pendingReq.Coins, &pendingReq.Rollno, &pendingReq.MadeAt)
		if er != nil {
			fmt.Fprint(w, "internal error")
			return
		}
		pendingReqs = append(pendingReqs, pendingReq)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pendingReqs)
}

func AcceptRejectRedeemRequest(w http.ResponseWriter, r *http.Request) {

	pl, err := isValidToken(r)

	exists, _ := Exists(pl.Rollno)

	if err != nil || !exists || !pl.IsAdmin {
		fmt.Fprint(w, "Not Authorised.")
		return
	}

	pendVer := &pendingVerdict{}
	json.NewDecoder(r.Body).Decode(pendVer)

	var status string
	err = Db.QueryRow("SELECT status FROM RedeemRequests WHERE requestID = ?", pendVer.RequestID).Scan(&status)
	if status != "p" {
		fmt.Fprint(w, "invalid request")
		return
	}
	if err != nil {
		fmt.Fprint(w, "action failed: please try again later")
		return
	}

	err = updateRedeemReq(pendVer)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, "action completed successfully")

}
