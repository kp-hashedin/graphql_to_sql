package main


import (
	"bufio"
    "fmt"
    "log"
    "os"
	"strings"
	"test_proj/config"
	"context"
	"time"
	"github.com/gomodule/redigo/redis"
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

	// Connecting to redis
	redis_conn, err := redis.Dial("tcp", "localhost:6379")
	if err!= nil {
		log.Fatal(err)
		defer redis_conn.Close()
	}
	fmt.Println("Connected with redis..")

    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)
    var text []string
	// `create table if not exists Home (
	// 	id serial not null primary key,
	// 	name varchar(255),
	// 	value double precision,
	// 	data date
	//  );`

	// alter table tweet  
    //   add column "user" text, 
    //   add constraint fk_test 
    //   foreign key ("user") 
    //   references "User" (id);

	// SELECT EXISTS (
	// 	SELECT FROM pg_tables
	// 	WHERE  schemaname = 'public'
	// 	AND    tablename  = 'tweet'
	// 	);

	m := make(map[string][]string)
	relation_map := make(map[string][]string)
	query := "create table if not exists "
	query_for_table_check := `SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename =  `
	sql_table_name := ""
	sql_column_name := ""
	var tablename_lc string
    for scanner.Scan() {
		temp := scanner.Text();

		// Table Name
		if (strings.Contains(temp, "type") && strings.Contains(temp, "struct")) {
			table_name := strings.Split(temp, " ")
			text = append(text, table_name[1])

			// check for Table name user(Reserved keyword)
			if ((table_name[1] == "User") || (table_name[1] == "USER") || (table_name[1] == "user")) {
				query = query + "\""+table_name[1]+"\"" + "( "
			} else {
				query = query + table_name[1] + "( "
			}
			sql_table_name = table_name[1]

			// entry for Exist check (required lower case)
			if table_name[1] != "User" {
				tablename_lc  = strings.ToLower(table_name[1])
				query_for_table_check = query_for_table_check +`'`+  tablename_lc+ `'` + ");"
			} else {
				query_for_table_check = query_for_table_check +`'`+  table_name[1]+ `'` + ");"
				tablename_lc = "\""+table_name[1]+"\""
			}

		} else if strings.Contains(temp, "json"){

			// Column Name
			temp1 := strings.Split(temp,  "`json:")
			x := temp1[1]
			
			if ((strings.Contains(temp, "int")) || (strings.Contains(temp, "string")) || (strings.Contains(temp, "[]") || (strings.Contains(temp, "bool")))) {
				query = query + x[1:len(x)-2]
				sql_column_name = x[1:len(x)-2]
			} else {
				sql_column_name = x[1:len(x)-2]
			}

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
			}  else if strings.Contains(y, "bool") {
				query = query + " BOOLEAN, "
			} else if strings.Contains(y, "*") && (!(strings.Contains(y, "int")) || (strings.Contains(y, "string")) || (strings.Contains(y, "[]") || (strings.Contains(y, "bool")))){
				// query = query + " JSON, "
				// fmt.Println(y)
				temp1Arr := strings.Split(y, "*")
				mappedField := temp1Arr[1]
				mappedField = strings.ReplaceAll(mappedField, " ", "")
				// Defining map with array of fields
				_, ok := m[sql_table_name] 
				_, okforRel := relation_map[sql_table_name]
				if (ok && okforRel) {
					list_of_fields := m[sql_table_name]
					list_of_fields = append(list_of_fields, sql_column_name)
					m[sql_table_name] = list_of_fields

					list_of_mappedWith := relation_map[sql_table_name]
					list_of_mappedWith = append(list_of_mappedWith, mappedField)
					relation_map[sql_table_name] = list_of_mappedWith
					
				} else {
					var list_of_fields []string
					list_of_fields = append(list_of_fields, sql_column_name)
					m[sql_table_name] = list_of_fields

					var list_of_mappedWith []string
					list_of_mappedWith = append(list_of_mappedWith, mappedField)
					relation_map[sql_table_name] = list_of_mappedWith
				}

			} else {}

		// Reseting the query and len variable for tracking Primary Key
		} else if strings.Contains(temp, "}"){
			query = query[:len(query)-2] + " );"
			if strings.Contains(query, "interface") {
				fmt.Println("Interface Not Allowed..")
			} else {
				ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)  
				defer cancelfunc()

				// Checking for Table existance
				rows,err := db.Query(query_for_table_check);
				if err != nil {
					log.Printf("Error -> %s", err)
					panic(err)
				}
				var value *bool
				for rows.Next() {
					if err := rows.Scan(&value); err != nil {
						log.Fatal(err)
					}
					if value == nil {
					}
				}
				if *value {
					// Drop Table then create
					query_drop := "DROP TABLE "
					query_drop_cascade := "DROP TABLE " 
					query_drop = query_drop + tablename_lc + ";"
					resDrop, err := db.ExecContext(ctx, query_drop)  
					fmt.Println(resDrop)
					if err != nil {  
						log.Printf("Error -> %s", err)
						query_drop_cascade = query_drop_cascade + tablename_lc + " CASCADE;"  
						resDrop, err = db.ExecContext(ctx, query_drop_cascade)
						if err != nil {
							log.Printf("Error -> %s", err)
							panic(err)
						}
					}
					log.Printf("Dropped Table -> %s", tablename_lc)
				}


				// Executung queries
				res, err := db.ExecContext(ctx, query)  
				fmt.Println(res)
				if err != nil {  
					log.Printf("Error -> %s", err)
					panic(err)
				}
			}
			query = "create table if not exists "
			sql_table_name = ""
			sql_column_name = ""
			tablename_lc =""
			query_for_table_check = `SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename =  `
		} else {}
    }
	fmt.Println("Migrated Successfully..")

	//ALter Table
	for index,element := range m{
		alterQuery := "alter table "
		x:= relation_map[index]
		alterQuery = alterQuery + checkReserveKeyword(index) 
		for id , record := range element{
			alterQuery = alterQuery+ " add column "
			target_element := checkReserveKeyword(record)
			target_relation_field := checkReserveKeyword(x[id])
			alterQuery = alterQuery + target_element +" text, add constraint fk"+ target_element  +" foreign key (" + target_element + ") references " + target_relation_field + " (id) ,"
		}
		alterQuery = alterQuery[:len(alterQuery)-1]
		alterQuery = alterQuery + ";"
		ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)  
		defer cancelfunc() 
		res, err := db.ExecContext(ctx, alterQuery)  
		fmt.Println(res)
		if err != nil {  
			log.Printf("Error -> %s", err)
		}
		alterQuery = "alter table "
    }
    file.Close()
}

func checkReserveKeyword(s string) string{
	if (s == "user") {
		return "\"user\""
	} else if (s == "User") {
		return "\"User\""
	} else if (s == "USER") {
		return "\"USER\""
	}
	return s
}
