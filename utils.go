package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func checkBalance(rollno string) (int, error) {

	row := Db.QueryRow("SELECT coins FROM User WHERE rollno = ?", rollno)
	var coinBalance int
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

	tx.Commit()
	return nil
}

func redeemCoins(red *redeem) error {

	tx, err := Db.Begin()
	if err != nil {
		return errors.New("transaction failed: something went wrong")
	}
	txRes, err2 := tx.Exec("UPDATE User SET coins = coins - ? WHERE rollno = ? AND coins >= ?", red.Coins, red.Rollno, red.Coins)
	if err2 != nil {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}
	rowsA, err3 := txRes.RowsAffected()
	if err3 != nil || rowsA != 1 {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	tx.Commit()
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
