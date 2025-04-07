package main

import (
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}

type CreateOrLoginResponse struct {
	Number int64  `json:"number"`
	Token  string `json:"token"`
}

type TransferRequest struct {
	ToAccount int64 `json:"toAccount"`
	Amount    int64 `json:"amount"`
}

type TransferResponse struct {
	Number     int64 `json:"number"`
	OldBalance int64 `json:"OldBalance"`
	NewBalance int64 `json:"NewBalance"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Password  string    `json:"password"`
	Number    int64     `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}

func (a *Account) validatePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password)) == nil
}

func NewAccount(firstName, lastName, password string) (*Account, error) {
	encpwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Password:  string(encpwd),
		Number:    int64(rand.Intn(10000000)),
		CreatedAt: time.Now().UTC(),
	}, nil
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

type claimsKey string
