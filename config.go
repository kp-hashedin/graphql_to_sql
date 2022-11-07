package config

import (
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
)

const (
	Host     = "HOSTNAME
	User     = "USERNAME"
	Password = "POASSWORD"
	Name     = "DBNAME"
	Port     = "5432"
)

func Setup() (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		Host,
		Port,
		User,
		Name,
		Password,
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
	panic(err)
	}
	// defer db.Close()
	return db, nil
}
