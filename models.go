package main

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type userDetails struct {
	Rollno   string `json:"rollno"`
	Name     string `json:"name"`
	Password string `json:"password"`
	// Batch    string `json:"batch"`
}

type userLogin struct {
	Rollno   string `json:"rollno"`
	Password string `json:"password"`
}

type tokenPayload struct {
	Rollno  string `json:"rollno"`
	IsAdmin bool   `json:"isAdmin"`
	// Batch   string `json:"batch"`
	jwt.StandardClaims
}

type recipient struct {
	Rollno string  `json:"rollno"`
	Coins  float32 `json:"coins"`
}

type transfer struct {
	// FromRollno -----Coins-----> ToRollno
	ToRollno   string  `json:"trollno"`
	FromRollno string  `json:"frollno"`
	Coins      float32 `json:"coins"`
	TaxedAmt   float32 `json:"tax"`
}

type redeem struct {
	Rollno string  `json:"rollno"`
	Coins  float32 `json:"coins"`
}
