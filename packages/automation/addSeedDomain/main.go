package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func Main(args map[string]interface{}) map[string]interface{} {
	//Get "domain" arg if it exists
	domain, ok := args["domain"].(string)
	domain = strings.ToLower(domain)
	if !ok {
		msg := make(map[string]interface{})
		msg["body"] = "Please send domain parameter!"
		return msg
	}

	platform, ok := args["platform"].(string)
	if !ok {
		msg := make(map[string]interface{})
		msg["body"] = "Please send platform parameter!"
		return msg
	}

	programUrl, ok := args["programUrl"].(string)
	if !ok {
		msg := make(map[string]interface{})
		msg["body"] = "Please send programUrl parameter!"
		return msg
	}

	if InsertSeedDomain(domain, platform, programUrl) {
		msg := make(map[string]interface{})
		msg["body"] = fmt.Sprintf("Added seed domain %s to program", domain)
		return msg
	} else {
		msg := make(map[string]interface{})
		msg["body"] = fmt.Sprintf("Seed domain %s is already in the database", domain)
		return msg
	}

}

func InsertSeedDomain(domain string, platform string, programUrl string) bool {
	//Check Connection
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Established a successful connection!")
	//Query
	sqlStatement := `INSERT INTO domains (domain, tested, platform, programurl) VALUES($1,FALSE,$2,$3);`
	_, err = db.Exec(sqlStatement, domain, platform, programUrl)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			fmt.Printf("seed domain %s already in db\n", domain)
			return false
		} else {
			panic(err)
		}
	} else {
		return true
	}
}
