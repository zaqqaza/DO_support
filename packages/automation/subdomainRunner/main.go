package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	_ "github.com/lib/pq"

	"crypto/tls"
	"strconv"

	"github.com/go-redis/redis"
)

type DomainObject struct {
	Domain string
}

var redisClient *redis.Client

func Main(args map[string]interface{}) map[string]interface{} {
	var domainsToQue []string
	var seedDomainList []string
	//Get "domain" arg if it exists
	domain, ok := args["domain"].(string)
	domain = strings.ToLower(domain)
	seedDomObj, err := QuerySeedDomains()
	if err != nil {
		panic(err)
	}
	for _, seedDomain := range seedDomObj {
		seedDomainList = append(seedDomainList, strings.ToLower(seedDomain.Domain))
	}
	//If "domain" arg was given, add only that domain to list
	//Otherwise add all seed domains to list
	if !ok {
		domainsToQue = seedDomainList
	} else {
		if contains(seedDomainList, domain) {
			domainsToQue = append(domainsToQue, domain)
		}
	}
	for _, d := range domainsToQue {
		PushToRedis(d)
	}

	//Sending message back to let user know how many domains where qued
	msg := make(map[string]interface{})
	msg["body"] = fmt.Sprintf("Queued %d domains from subdomainRunner", len(domainsToQue))
	return msg
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func QuerySeedDomains() ([]DomainObject, error) {
	var DomainObjects []DomainObject
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
	rows, err := db.Query("SELECT domain FROM domains")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var dobj DomainObject
		if err := rows.Scan(&dobj.Domain); err != nil {
			panic(err)
		}
		DomainObjects = append(DomainObjects, dobj)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return DomainObjects, nil
}

func PushToRedis(domain string) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:      os.Getenv("REDIS_ADDR"),
		Password:  os.Getenv("REDIS_PASS"),
		DB:        0,                                     // redis databases are deprecated so we will just use the default
		TLSConfig: &tls.Config{InsecureSkipVerify: true}, // needed for the standard DO redis
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		fmt.Println("Unable to connect to specified Redis server:", err)
		os.Exit(1)
	}

	queueID := uuid.New().String()
	redisClient.LPush("subdomain-worker", queueID+":::_:::"+strconv.Itoa(50000)+":::_:::"+domain)
}
