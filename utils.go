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

func addUser(db *sql.DB, usr *userDetails) error {
	row := db.QueryRow("SELECT COUNT(*) FROM User WHERE rollno = ?", usr.Rollno)
	var has bool
	row.Scan(&has)
	if !has {
		stmt, _ := db.Prepare("INSERT INTO User (rollno, name, password) VALUES (?, ?, ?)")
		_, err := stmt.Exec(usr.Rollno, usr.Name, usr.Password)

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

		token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // check if the tokenString is in the correct format
				return nil, fmt.Errorf("internal error")
			}
			return []byte(jwtSignature), nil
		})

		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"message": "Not Authorised"})
		}

		if token.Valid {
			handler(w, r)
		}

	})
}
