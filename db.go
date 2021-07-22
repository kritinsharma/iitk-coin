package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var Db *sql.DB

var redisClient *redis.Client

func Exists(rollNo string) (bool, error) {
	var has bool
	err := Db.QueryRow("SELECT COUNT(*) FROM User WHERE rollno = ?", rollNo).Scan(&has)
	if err != nil {
		return false, errors.New("internal error")
	}
	return has, nil
}

func AdminFlag(rollNo string) (bool, error) {
	var isAdmin bool
	err := Db.QueryRow("SELECT isAdmin FROM User WHERE rollno = ?", rollNo).Scan(&isAdmin)
	if err == sql.ErrNoRows {
		return false, errors.New("user with given rollno does not exist")
	}
	if err != nil {
		return false, errors.New("internal error")
	}
	return isAdmin, nil
}

func getBatch(rollNo string) (string, error) {
	b := rollNo[:2]
	return b, nil
}

func addUser(usr *userDetails) error {

	has, err := Exists(usr.Rollno)

	if err != nil {
		return err
	}

	if !has {
		_, err := Db.Exec("INSERT INTO User (rollno, name, password, coins, isAdmin, email) VALUES (?, ?, ?, ?, ?, ?)", usr.Rollno, usr.Name, usr.Password, 0, usr.Rollno == "000000", usr.Email)

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

func createRedeemRequest(red *redeem) error {
	_, err := Db.Exec("INSERT INTO RedeemRequests (itemName, coins, madeBy) VALUES (?, ?, ?)", red.ItemName, red.Coins, red.Rollno)
	if err != nil {
		return errors.New("internal error: something went wrong")
	}
	return nil
}
