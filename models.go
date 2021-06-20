package main

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type userDetails struct {
	Rollno   string `json:"rollno"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type userLogin struct {
	Rollno   string `json:"rollno"`
	Password string `json:"password"`
}

type token struct {
	Rollno string `json:"rollno"`
	jwt.StandardClaims
}

type coinBal struct {
	Rollno string `json:"rollno"`
}

type recipient struct {
	Rollno string `json:"rollno"`
	Coins  int    `json:"coins"`
}

type transaction struct {
	// FromRollno -----Coins-----> ToRollno
	ToRollno   string `json:"trollno"`
	FromRollno string `json:"frollno"`
	Coins      int    `json:"coins"`
}
