package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	gomail "gopkg.in/gomail.v2"
)

var ctx = context.TODO()

func checkBalance(rollno string) (float32, error) {

	row := Db.QueryRow("SELECT coins FROM User WHERE rollno = ?", rollno)
	var coinBalance float32
	err := row.Scan(&coinBalance)

	if err == sql.ErrNoRows {
		return -1, errors.New("invalid roll number")
	}
	if err != nil {
		return -1, errors.New("something went wrong")
	}

	return coinBalance, nil
}

func addCoins(rec *recipient) error {

	tx, err := Db.Begin()
	if err != nil {
		return errors.New("transaction failed: something went wrong")
	}

	txRes, err2 := tx.Exec("UPDATE User SET coins = coins + ? WHERE rollno = ? AND coins + ? <= 1000", rec.Coins, rec.Rollno, rec.Coins)
	if err2 != nil {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	rowsA, err3 := txRes.RowsAffected()
	if err3 != nil || rowsA != 1 {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	txRes, err2 = tx.Exec("INSERT INTO Transaction_Logs (mode, secondaryUser, deb_cred_secondary) VALUES (?, ?, ?)", "REWARD", rec.Rollno, rec.Coins)
	if err2 != nil {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	rowsA, err3 = txRes.RowsAffected()
	if err3 != nil || rowsA != 1 {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	tx.Commit()
	return nil

}

func sendCoins(trf *transfer) error {

	tx, err := Db.Begin()
	if err != nil {
		return errors.New("transaction failed: something went wrong")
	}
	txRes, err2 := tx.Exec("UPDATE User SET coins = coins - ? WHERE rollno = ? AND coins >= ?", trf.Coins, trf.FromRollno, trf.Coins)
	if err2 != nil {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	rowsA, err3 := txRes.RowsAffected()
	if err3 != nil || rowsA != 1 {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	txRes, err2 = tx.Exec("UPDATE User SET coins = coins + ? WHERE rollno = ? AND coins + ? <= 1000", trf.TaxedAmt*trf.Coins, trf.ToRollno, trf.TaxedAmt*trf.Coins)
	if err2 != nil {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	rowsA, err3 = txRes.RowsAffected()
	if err3 != nil || rowsA != 1 {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	txRes, err2 = tx.Exec("INSERT INTO Transaction_Logs (mode, secondaryUser, deb_cred_secondary, primaryUser, deb_cred_primary) VALUES (?, ?, ?, ?, ?)", "TRANSFER", trf.ToRollno, trf.TaxedAmt*trf.Coins, trf.FromRollno, -trf.Coins)
	if err2 != nil {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	rowsA, err3 = txRes.RowsAffected()
	if err3 != nil || rowsA != 1 {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	tx.Commit()
	return nil
}

func updateRedeemReq(pv *pendingVerdict) error {

	tx, err := Db.Begin()
	if err != nil {
		return errors.New("action failed: something went wrong")
	}
	txRes, err2 := tx.Exec("UPDATE RedeemRequests SET status = ? WHERE requestID = ?", pv.Verdict, pv.RequestID)

	if err2 != nil {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}
	rowsA, err3 := txRes.RowsAffected()
	if err3 != nil || rowsA != 1 {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	if pv.Verdict == "a" {
		var red redeem
		err = tx.QueryRow("SELECT madeBy, coins FROM RedeemRequests WHERE requestID = ?", pv.RequestID).Scan(&red.Rollno, &red.Coins)
		if err != nil {
			tx.Rollback()
			return errors.New("transaction failed: something went wrong")
		}
		txRes, err2 = tx.Exec("UPDATE User SET coins = coins - ? WHERE rollno = ? AND coins >= ?", red.Coins, red.Rollno, red.Coins)
		if err2 != nil {
			tx.Rollback()
			return errors.New("transaction failed: something went wrong")
		}
		rowsA, err3 = txRes.RowsAffected()
		if err3 != nil || rowsA > 1 {
			tx.Rollback()
			return errors.New("transaction failed: something went wrong")
		}
		if rowsA < 1 {
			tx.Rollback()
			return errors.New("balance not enough")
		}
	}
	tx.Commit()
	return nil

}

func mailOTP(rollNo string) error {

	otp := make([]byte, 6)
	_, err := rand.Read(otp)
	if err != nil {
		return errors.New("something went wrong")
	}
	OTP := make([]rune, 6)
	for i := 0; i < 6; i += 1 {
		OTP[i] = rune(int(otp[i])%10 + '0')
	}
	var mailto string
	err = Db.QueryRow("SELECT email FROM User WHERE rollno = ?", rollNo).Scan(&mailto)
	if err != nil {
		return errors.New("something went wrong")
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	m := gomail.NewMessage()

	m.SetHeader("From", fromEmail)
	m.SetHeader("To", mailto)
	m.SetHeader("Subject", "OTP from iitk-coin")
	m.SetBody("text/HTML", fmt.Sprintf("Your otp is %s.<br>It will expire in 3 minutes<br>Do not share your otp with anyone else.", string(OTP)))

	dialer := gomail.NewDialer(smtpHost, smtpPort, fromEmail, emailPasswd)

	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err = dialer.DialAndSend(m)
	if err != nil {
		return errors.New("something went wrong")
	}
	err = redisClient.Set(ctx, rollNo, string(OTP), 3*time.Minute).Err()
	if err != nil {
		return errors.New("something went wrong")
	}
	return nil
}

func GetToken(rollno string, isAdmin bool, batch string) (string, error) {

	expiresAt := time.Now().Add(time.Minute * 30).Unix()

	tkn := &tokenPayload{
		Rollno:  rollno,
		IsAdmin: isAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tkn)
	tokenString, err := token.SignedString([]byte(jwtSignature))

	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func isValidToken(r *http.Request) (tokenPayload, error) {

	if r.Header["Token"] == nil {
		return tokenPayload{}, errors.New("no token")
	}

	payload := tokenPayload{}

	token, err := jwt.ParseWithClaims(r.Header["Token"][0], &payload, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // check if the tokenString is in the correct format
			return nil, fmt.Errorf("internal error")
		}
		return []byte(jwtSignature), nil
	})

	if err != nil {
		return tokenPayload{}, errors.New("could not parse token")
	}

	if token.Valid {
		return payload, nil
	}

	return tokenPayload{}, errors.New("token expired")
}
