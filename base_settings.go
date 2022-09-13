package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"strings"
)

const (
	username = ""
	password = ""
	host = "localhost"
	port = "5433"
	db_name = ""
)

func ConnectingToTheBase () *pgxpool.Pool{
	// Connecting to the base
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, db_name)
	// Initializing the connection cursor
	cursor, err := pgxpool.Connect(context.Background(), databaseUrl)

	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
	return cursor
}

func CreateBaseOrDoNothing (cursor *pgxpool.Pool) {
	// Create a table if it doesn't exist

	_, err := cursor.Exec(context.Background(),
		"create table if not exists Companies" +
			" (company_scope TEXT NULL, " +
			"name TEXT NULL, " +
			"cik TEXT NULL, " +
			"sic TEXT NULL, " +
			"phone TEXT NULL, " +
			"business_street TEXT NULL, " +
			"business_city TEXT NULL, " +
			"business_state TEXT NULL, " +
			"business_zip TEXT NULL, " +
			"main_street TEXT NULL, " +
			"main_city TEXT NULL,  " +
			"main_state TEXT NULL, " +
			"main_zip TEXT NULL) ")


	_, err = cursor.Exec(context.Background(),
		"create table if not exists sic_codes " +
			"(division TEXT NULL, " +
			"major_group TEXT NULL, " +
			"industry_group TEXT NULL, " +
			"sic TEXT NULL, " +
			"description TEXT NULL) ")

	if err != nil {
		log.Fatal("Create Base failed: ", err)
	}
}

func AddingDataToDb (tags map[string]string, tagsAddress map[string][]string, companyScope string, cursor *pgxpool.Pool) {
	// Adding data to the database, the number of columns in the table is 12
	_, err := cursor.Exec(context.Background(), `INSERT INTO Companies 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`, companyScope,
		tags["<CONFORMED-NAME>"], tags["<CIK>"], tags["<ASSIGNED-SIC>"], tags["<PHONE>"],
		tagsAddress["<STREET1>"][0], tagsAddress["<CITY>"][0], tagsAddress["<STATE>"][0], tagsAddress["<ZIP>"][0],
		tagsAddress["<STREET1>"][1], tagsAddress["<CITY>"][1], tagsAddress["<STATE>"][1], tagsAddress["<ZIP>"][1],
	)


	if err != nil {
		log.Fatal( "AddingDataDb failed: ", err)
	}
}

func CheckExistCik (basicInformation []string, cursor *pgxpool.Pool) bool {
	// Check if the record is already in the database, search by cik (unique identifier)

	checkExistCik := basicInformation[1]
	checkExistCik = strings.Replace(checkExistCik, "CIK>", "", 1)

	var status bool
	rows, _ := cursor.Query(context.Background(), fmt.Sprintf("SELECT EXISTS (SELECT * FROM Companies WHERE cik ='%s');", checkExistCik))

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&status)
		if err != nil {
			log.Fatal("CheckExistCik failed: ", err)
		}
	}

	return status

}

// Section that works with the database of SIC codes

func GetCompanyScope (cursor *pgxpool.Pool, sic string) string{
	companyScope := ""
	if !(sic == "0000") && !(len(sic) == 0) {
		err := cursor.QueryRow(context.Background(), "SELECT description FROM sic_codes WHERE sic=$1", sic).Scan(&companyScope)

		if err != nil {
			if err == pgx.ErrNoRows {
				industrialGroup := sic[0:2]
				err = cursor.QueryRow(context.Background(), "SELECT description FROM sic_codes WHERE industry_group=$1", industrialGroup).Scan(&companyScope)
			} else {
				log.Fatal("GetCompanyScope: ", err)
			}
		}
	}

	return companyScope

}


type sicCodes struct {
	division string
	majorGroup string
	industryGroup string
	sic string
	description string

}


func AddDataCSV (cursor *pgxpool.Pool) {
	// Creates a base from a CSV file containing sic
	// and its corresponding description

	var count int
	err := cursor.QueryRow(context.Background(), `SELECT count(*) FROM sic_codes`).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		records, err := readData("sic-codes.csv")

		if err != nil {
			log.Fatal(err)
		}

		for _, record := range records {

			data := sicCodes{
				division:  record[0],
				majorGroup: record[1],
				industryGroup:   record[2],
				sic: record[3],
				description: record[4],
			}
			_, err := cursor.Exec(context.Background(), `INSERT INTO sic_codes 
		VALUES ($1, $2, $3, $4, $5);`, data.division, data.majorGroup, data.industryGroup, data.sic, data.description,
			)

			if err != nil {
				log.Fatal("AddDataCSV failed: ", err)
			}
		}
	}


}


func readData(fileName string) ([][]string, error) {
	// Adding data to a CSV file

	f, err := os.Open(fileName)

	if err != nil {
		return [][]string{}, err
	}

	defer f.Close()

	r := csv.NewReader(f)

	// skip first line
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}


// The section where the work with logs takes place

func CreateLogs (cursor *pgxpool.Pool) {

	_, err := cursor.Exec(context.Background(),
		"create table if not exists logs_data " +
		"(logs TEXT NULL)")

	if err != nil {
		log.Fatal ("create logs-database failed: ", err)
	}

}

func GetLogs (cursor *pgxpool.Pool) string{
	var rawData string

	err := cursor.QueryRow(context.Background(), "SELECT logs FROM logs_data LIMIT 1").Scan(&rawData)

	if err != nil {
		if err == pgx.ErrNoRows {
		} else {
			log.Fatal ("Getting logs failed: ", err)
	}
	}

	return rawData
}

func SendLogs (cursor *pgxpool.Pool, json []byte) {
	_, err := cursor.Exec(context.Background(), `DELETE FROM logs_data`)

	if err != nil {
		log.Fatal ("clean logs failed: ", err)
	}

	_, err = cursor.Exec(context.Background(), `INSERT INTO logs_data VALUES ($1)`, json)

	if err != nil {
		log.Fatal ("add logs failed: ", err)
	}
}
