package config

import (
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
)

const (
	Host     = "satao.db.elephantsql.com"
	User     = "bbidheld"
	Password = "hjXHmHIpHSxO2crbANFE26gmYkhml1nI"
	Name     = "bbidheld"
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
