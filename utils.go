package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func addUser(usr *userDetails) error {
	row := Db.QueryRow("SELECT COUNT(*) FROM User WHERE rollno = ?", usr.Rollno)
	var has bool
	row.Scan(&has)
	if !has {
		stmt, _ := Db.Prepare("INSERT INTO User (rollno, name, password, coins) VALUES (?, ?, ?, ?)")
		_, err := stmt.Exec(usr.Rollno, usr.Name, usr.Password, 0)

		if err != nil {
			return errors.New("could not add user: something went wrong")
		} else {
			fmt.Printf("User [%s: %s] succesfully added.\n", usr.Rollno, usr.Name)
			return nil
		}

	} else {
		return fmt.Errorf("user with roll number %v already present", usr.Rollno)
	}
}

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

func sendCoins(trx *transaction) error {

	tx, err := Db.Begin()
	if err != nil {
		return errors.New("transaction failed: something went wrong")
	}

	txRes, err2 := tx.Exec("UPDATE User SET coins = coins - ? WHERE rollno = ? AND coins >= ?", trx.Coins, trx.FromRollno, trx.Coins)
	if err2 != nil {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	rowsA, err3 := txRes.RowsAffected()
	if err3 != nil || rowsA != 1 {
		tx.Rollback()
		return errors.New("transaction failed: something went wrong")
	}

	txRes, err2 = tx.Exec("UPDATE User SET coins = coins + ? WHERE rollno = ? AND coins + ? <= 1000", trx.Coins, trx.ToRollno, trx.Coins)
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

func GetToken(rollno string) (string, error) {

	expiresAt := time.Now().Add(time.Minute * 30).Unix()

	tkn := &token{
		Rollno: rollno,
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

func isAuthorised(handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Token"] == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"message": "Not Authorised"})
			return
		}

		token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // check if the tokenString is in the correct format
				return nil, fmt.Errorf("internal error")
			}
			return []byte(jwtSignature), nil
		})

		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"message": "Not Authorised"})
			return
		}

		if token.Valid {
			handler(w, r)
		}

	})
}
