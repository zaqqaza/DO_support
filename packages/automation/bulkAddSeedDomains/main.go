package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type program struct {
	Name    string
	Url     string
	Bounty  bool
	Domains []string
}

type Data struct {
	Programs []program
}

var domCount int

func Main(args map[string]interface{}) map[string]interface{} {
	req, err := http.NewRequest("GET", "https://raw.githubusercontent.com/projectdiscovery/public-bugbounty-programs/main/chaos-bugbounty-list.json", nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var payload Data
	err = json.Unmarshal(b, &payload)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	domCount = 0
	for _, prog := range payload.Programs {
		if prog.Bounty {
			if strings.Contains(prog.Url, "hackerone.com") {
				addDomains("Hackerone", prog.Url, prog.Domains)
			} else if strings.Contains(prog.Url, "bugcrowd.com") {
				addDomains("Bugcrowd", prog.Url, prog.Domains)
			} else if strings.Contains(prog.Url, "immunefi.com") {
				addDomains("Immunefi", prog.Url, prog.Domains)
			} else if strings.Contains(prog.Url, "intigriti.com") {
				addDomains("Intigriti", prog.Url, prog.Domains)
			} else if strings.Contains(prog.Url, "yeswehack.com") {
				addDomains("Yeswehack", prog.Url, prog.Domains)
			}
		}
	}
	msg := make(map[string]interface{})
	msg["body"] = fmt.Sprintf("Added %d seed domains to automation", domCount)
	return msg
}

func addDomains(platform string, programUrl string, domains []string) {
	for _, domain := range domains {
		if InsertSeedDomain(domain, platform, programUrl) {
			domCount++
		} else {
			fmt.Println(domain + "already is in seed domain DB")
		}
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
