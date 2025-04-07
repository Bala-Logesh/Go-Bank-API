package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/", handleRoot)

	router.HandleFunc("/login", makeHTTPHandlerFunc(s.handleLogin))
	router.HandleFunc("/account", makeHTTPHandlerFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandlerFunc(s.handleAccountWithID), s.store))
	router.HandleFunc("/deposit/{id}", withJWTAuth(makeHTTPHandlerFunc(s.handleDepositMoney), s.store))
	router.HandleFunc("/withdraw/{id}", withJWTAuth(makeHTTPHandlerFunc(s.handleWithdrawMoney), s.store))
	router.HandleFunc("/transfer/{id}", withJWTAuth(makeHTTPHandlerFunc(s.handleTransfer), s.store))

	log.Println("JSON API server running on port", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello there from server!!")
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleAccountWithID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccountByID(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, _ *http.Request) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)

	if err != nil {
		return err
	}

	account, err := s.store.GetAccountByID(id)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	acc, err := s.store.GetAccountByNumber(req.Number)

	if err != nil {
		return err
	}

	if !acc.validatePassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}

	tokenString, err := createJWT(acc)
	if err != nil {
		return err
	}

	fmt.Println("JWT Token: ", tokenString)

	resp := CreateOrLoginResponse{
		Number: acc.Number,
		Token:  tokenString,
	}

	WriteJSON(w, http.StatusOK, resp)

	return nil
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	// Pointer approach
	// createAccountReq := CreateAccountRequest{}
	// if err := json.NewDecoder(r.Body).Decode(&createAccountReq); err != nil {
	// 	return err
	// }

	// Object approach
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}

	account, err := NewAccount(createAccountReq.FirstName, createAccountReq.LastName, createAccountReq.Password)

	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}

	fmt.Println("JWT Token: ", tokenString)

	resp := CreateOrLoginResponse{
		Number: account.Number,
		Token:  tokenString,
	}

	return WriteJSON(w, http.StatusOK, resp)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)

	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *APIServer) handleDepositMoney(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Amount int64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return fmt.Errorf("invalid amount given: %v", err)
	}

	number := getAccountFromCxt(r)
	acc, err := s.store.GetAccountByNumber(number)

	if err != nil {
		return err
	}

	oldBalance := acc.Balance

	defer r.Body.Close()

	newBalance, err := modifyAccountBalance(s.store, acc, body.Amount)

	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, ApiError{Error: "Something went wrong"})
		return err
	}

	transferResp := TransferResponse{
		Number:     number,
		OldBalance: oldBalance,
		NewBalance: newBalance,
	}

	return WriteJSON(w, http.StatusOK, transferResp)
}

func (s *APIServer) handleWithdrawMoney(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Amount int64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return fmt.Errorf("invalid amount given: %v", err)
	}

	number := getAccountFromCxt(r)
	acc, err := s.store.GetAccountByNumber(number)

	if err != nil {
		return err
	}

	oldBalance := acc.Balance

	defer r.Body.Close()

	if oldBalance < body.Amount {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "Insufficient Balance"})
	}

	newBalance, err := modifyAccountBalance(s.store, acc, -1*body.Amount)

	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, ApiError{Error: "Something went wrong"})
		return err
	}

	transferResp := TransferResponse{
		Number:     number,
		OldBalance: oldBalance,
		NewBalance: newBalance,
	}

	return WriteJSON(w, http.StatusOK, transferResp)
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)

	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}

	number := getAccountFromCxt(r)
	fromAccount, err := s.store.GetAccountByNumber(number)

	if err != nil {
		return err
	}

	defer r.Body.Close()

	oldBalance := fromAccount.Balance

	if fromAccount.Balance < transferReq.Amount {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "Insufficient Balance"})
	}

	if fromAccount.Number == int64(transferReq.ToAccount) {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "Cannot transfer money to the same account"})
	}

	toAccount, err := s.store.GetAccountByNumber(transferReq.ToAccount)

	if err != nil {
		return err
	}

	newBalance, err := modifyAccountBalance(s.store, fromAccount, -1*transferReq.Amount)

	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, ApiError{Error: "Something went wrong"})
	}

	_, err = modifyAccountBalance(s.store, toAccount, transferReq.Amount)

	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, ApiError{Error: "Something went wrong"})
	}

	transferResp := TransferResponse{
		Number:     fromAccount.Number,
		OldBalance: oldBalance,
		NewBalance: newBalance,
	}

	return WriteJSON(w, http.StatusOK, transferResp)
}

// Helper functions
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}

	secret := getJWTSecret()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Calling JWT auth middleware")
		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)

		if err != nil {
			permissionDenied(w)
			return
		}

		if !token.Valid {
			permissionDenied(w)
			return
		}

		userID, err := getID(r)

		if err != nil {
			fmt.Println(err)
			permissionDenied(w)
			return
		}

		account, err := s.GetAccountByID(userID)

		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int64(claims["accountNumber"].(float64)) {
			permissionDenied(w)
			return
		}

		claimskey := claimsKey("accountNumber")

		ctx := context.WithValue(r.Context(), claimskey, account.Number)

		handlerFunc(w, r.WithContext(ctx))
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := getJWTSecret()

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

// This method decorates the functions that we have which are of the type apiFunc into type http.Handler that the http HandleFunc expects
func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			// handle the error
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) == 0 {
		log.Fatalf("JWT_SECRET environment variable is not set")
	}

	return secret
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)

	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}

	return id, nil
}

func getAccountFromCxt(r *http.Request) int64 {
	claimsKey := claimsKey("accountNumber")

	number := r.Context().Value(claimsKey).(int64)

	return number
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "Permission Denied"})
}

func modifyAccountBalance(s Storage, acc *Account, amount int64) (int64, error) {
	acc.Balance += amount

	err := s.UpdateAccountBalance(acc)

	if err != nil {
		return 0, err
	}

	return acc.Balance, nil
}