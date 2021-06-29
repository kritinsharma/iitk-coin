package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

var Db *sql.DB

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
		_, err := Db.Exec("INSERT INTO User (rollno, name, password, coins, isAdmin) VALUES (?, ?, ?, ?, ?)", usr.Rollno, usr.Name, usr.Password, 0, 0)

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

/*TransactionID INT PRIMARY KEY AUTO_INCREMENT,
mode TEXT,
primaryUser TEXT,
deb_cred_primary FLOAT,
secondaryUser TEXT,
deb_cred_secondary FLOAT,
madeAt DATE DEFAULT (DATETIME('now', 'localtime')),
status BOOL*/

func logTransfer(trf *transfer, status bool) {
	_, err := Db.Exec("INSERT INTO Transaction_Logs (mode, secondaryUser, deb_cred_secondary, primaryUser, deb_cred_primary, status) VALUES (?, ?, ?, ?, ?, ?)", "TRANSFER", trf.ToRollno, trf.TaxedAmt*trf.Coins, trf.FromRollno, -trf.Coins, status)
	if err != nil {
		log.Fatal(err)
	}
}

func logReward(rec *recipient, status bool) {
	_, err := Db.Exec("INSERT INTO Transaction_Logs (mode, secondaryUser, deb_cred_secondary, status) VALUES (?, ?, ?, ?)", "REWARD", rec.Rollno, rec.Coins, status)
	if err != nil {
		log.Fatal(err)
	}
}

func logRedeem(red *redeem, status bool) {
	_, err := Db.Exec("INSERT INTO Transaction_Logs (mode, primaryUser, deb_cred_primary, status) VALUES (?, ?, ?, ?)", "REDEEM", red.Rollno, -red.Coins, status)
	if err != nil {
		log.Fatal(err)
	}
}
