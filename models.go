package main

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type userDetails struct {
	Rollno   string `json:"rollno"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userLogin struct {
	Rollno   string `json:"rollno"`
	Password string `json:"password"`
}

type tokenPayload struct {
	Rollno  string `json:"rollno"`
	IsAdmin bool   `json:"isAdmin"`
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
	OTP        string  `json:"otp"`
}

type redeem struct {
	Rollno   string  `json:"rollno"`
	Coins    float32 `json:"coins"`
	ItemName string  `json:"itemName"`
	OTP      string  `json:"otp"`
}

type redeemRequest struct {
	RequestID int    `json:"requestID"`
	Rollno    string `json:"rollno"`
	Coins     string `json:"coins"`
	ItemName  string `json:"itemName"`
	MadeAt    string `json:"madeAt"`
	Status    string `json:"status"`
}

type pendingVerdict struct {
	RequestID int    `json:"requestID"`
	Verdict   string `json:"verdict"`
}
