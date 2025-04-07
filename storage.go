package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account, string, string, string) error
	UpdateAccountBalance(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(int64) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	log.Println("Connecting to database")
	connStr := os.Getenv("POSTGRES_URL")
	if len(connStr) == 0 {
		log.Fatalf("POSTGRES_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to Postgres DB")

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	s.CreateAccountTable()
	return nil
}

func (s *PostgresStore) CreateAccountTable() error {
	query := `create table if not exists account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		password varchar(255),
		number serial,
		balance serial,
		created_at timestamp
	)`

	_, err := s.db.Exec(query)

	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := `insert into account (
		first_name, last_name, password, number, balance, created_at
	) values (
		$1, $2, $3, $4, $5, $6
	)`

	resp, err := s.db.Exec(query, acc.FirstName, acc.LastName, acc.Password, acc.Number, acc.Balance, acc.CreatedAt)

	if err != nil {
		return err
	}

	fmt.Printf("Account created: %+v\n", resp)

	return nil
}

func (s *PostgresStore) UpdateAccount(acc *Account, firstName, lastName, password string) error {
	setClauses := []string{}
	args := []interface{}{}
	argID := 1

	if firstName != "" {
		setClauses = append(setClauses, fmt.Sprintf("first_name = $%d", argID))
		args = append(args, firstName)
		argID++
	}
	if lastName != "" {
		setClauses = append(setClauses, fmt.Sprintf("last_name = $%d", argID))
		args = append(args, lastName)
		argID++
	}
	if password != "" {
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", argID))
		args = append(args, password)
		argID++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`UPDATE account SET %s WHERE id = $%d`,
		strings.Join(setClauses, ", "), argID)

	args = append(args, acc.ID)

	_, err := s.db.Exec(query, args...)
	return err
}

func (s *PostgresStore) UpdateAccountBalance(acc *Account) error {
	query := `update account set
		balance = $1
	where id = $2`

	resp, err := s.db.Exec(query, acc.Balance, acc.ID)

	if err != nil {
		return err
	}

	fmt.Printf("Account Balance updated: %+v\n", resp)

	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	query := "delete from account where id = $1"

	_, err := s.db.Query(query, id)

	return err
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	query := "select * from account"

	rows, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for rows.Next() {
		account, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	query := "select * from account where id = $1"

	rows, err := s.db.Query(query, id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) GetAccountByNumber(number int64) (*Account, error) {
	query := "select * from account where number = $1"

	rows, err := s.db.Query(query, number)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found", number)
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(&account.ID, &account.FirstName, &account.LastName, &account.Password, &account.Number, &account.Balance, &account.CreatedAt)

	return account, err
}
