package main

import (
	"bufio"
    "fmt"
    "log"
    "os"
	"strings"
	"grapql_to_sql/config"
	"context"
	"time"
)

func main () {
	db, err := config.Setup()
	if err !=nil {
		log.Panic(err)
		return
	}
	fmt.Println("Connected with Database...")

	file, err := os.Open("graph/model/models_gen.go")
	if err != nil {
        log.Fatalf("failed to open")
  
    }
    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)
    var text []string
	// `create table if not exists Home (
	// 	id serial not null primary key,
	// 	name varchar(255),
	// 	value double precision,
	// 	data date
	//  );`
	query := "create table if not exists "
    for scanner.Scan() {
		temp := scanner.Text();

		// Table Name
		if strings.Contains(temp, "type") {
			table_name := strings.Split(temp, " ")
			text = append(text, table_name[1])
			query = query + table_name[1] + "( "

		} else if strings.Contains(temp, "json"){

			// Column Name
			temp1 := strings.Split(temp,  "`json:")
			x := temp1[1]
			
			query = query + x[1:len(x)-2]
			
			// Data Type
			y := temp1[0]
			if strings.Contains(y, "int") {
				if (!(strings.Contains(query, "NUMERIC") ||  strings.Contains(query, "TEXT") ||  strings.Contains(query, "JSON") || strings.Contains(query, "BOOLEAN")))  {
					query = query + " NUMERIC PRIMARY KEY, "
				} else {
					query = query + " NUMERIC, "
				}
			} else if strings.Contains(y, "string") {
				if (!(strings.Contains(query, "NUMERIC") ||  strings.Contains(query, "TEXT") ||  strings.Contains(query, "JSON")  || strings.Contains(query, "BOOLEAN")))  {
					query = query + " TEXT PRIMARY KEY, "
				} else {
					query = query + " TEXT, "
				}
			} else if strings.Contains(y, "[]*"){
				query = query + " JSON[], "
			} else if strings.Contains(y, "*") && (!(strings.Contains(y, "int")) || (strings.Contains(y, "string")) || (strings.Contains(y, "[]") || (strings.Contains(y, "bool")))){
				query = query + " JSON, "
			} else if strings.Contains(y, "bool") {
				query = query + " BOOLEAN, "
			} else {}

		// Reseting the query and len variable for tracking Primary Key
		} else if strings.Contains(temp, "}"){
			fmt.Println(" ")
			fmt.Println(" ")
			query = query[:len(query)-2] + " );"
			fmt.Println(query)
			if strings.Contains(query, "interface") {
				fmt.Println("Interface Not Allowed..")
			} else {
				ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)  
				defer cancelfunc() 
				res, err := db.ExecContext(ctx, query)  
				fmt.Println(res)
				if err != nil {  
					log.Printf("Error %s when creating product table", err)
				}
			}
			query = "create table if not exists "
		} else {}
    }
	fmt.Println("Migrated Successfully..")
    file.Close()
}