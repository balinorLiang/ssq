package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=127.0.0.1 port=5432 user=ivorysql password=123456 dbname=ivorysql sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("error: db open error")
		return
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("error: db connection error")
		return
	}

	startNum := 0
	currentIssue := 2024114
	err = db.QueryRow("select max(issue) from ssq_history").Scan(&startNum)
	if err != nil {
		startNum = 2024050
	}
	fmt.Println(startNum)
	for i := 0; i < currentIssue-startNum; i++ {
		num, issue, time, err := getSsqResultFromURL(startNum + i)
		if err != nil {
			break
		}

		numList := strings.Split(num, "|")
		err = saveSsqResult(db, issue, numList, time)
		if err != nil {
			return
		}
	}

}

func saveSsqResult(db *sql.DB, issue string, numList []string, time string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("insert into ssq_history (issue, num1, num2, num3, num4, num5, num6, refnum, time) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)", issue, numList[0], numList[1], numList[2], numList[3], numList[4], numList[5], numList[6], time)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func getSsqResultFromURL(issue int) (string, string, string, error) {
	fullURL := fmt.Sprintf("%s?id=%s&key=%s&qh=%d", URL, id, key, issue)
	resp, err := http.Post(fullURL, "", nil)
	if err != nil {
		fmt.Println(err)
		return "", "", "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", "", "", err
	}

	defer resp.Body.Close()

	result := SsqResult{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", "", "", err
	}

	if resp.StatusCode != 200 {
		return "", "", "", errors.New("status code wrong")
	}

	if result.Number == "" {
		time.Sleep(1 * time.Minute)
		fmt.Println("调用次数过快，稍等", string(body))
		return getSsqResultFromURL(issue)
	}

	fmt.Println("result:", resp.StatusCode, result.Number, result.Issue, result.Time)

	return result.Number, result.Issue, result.Time, nil
}
