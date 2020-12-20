package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"github.com/theplant/luhn"
)

/*
	{
		ccData : {
			num: string,
			expiry : string,
			cvv: string
		},
		moneyData : {
			currency : string,
			amount : string // money as a double is bad
		}
	}
*/

// Auth is the method assigned to /authorize
func Auth(w http.ResponseWriter, r *http.Request) {
	var parsedBody InputData //marshal the json into a nice struct
	var db *badger.DB

	wGroup := sync.WaitGroup{}
	wGroup.Add(4)
	errChan := make(chan error, 4)

	file, fileErr := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if fileErr == nil {
		defer file.Close()
		log.SetOutput(file)
	}

	err := json.NewDecoder(r.Body).Decode(&parsedBody)

	if err != nil {
		log.Fatal(err)
	}

	atoiCC, _ := strconv.Atoi(parsedBody.CcData.Num)
	uuid, _ := uuid.New().MarshalText()
	uuidStr := string(uuid)

	// 'edge' case
	go checkEdgeCaseAuth(w, uuidStr, parsedBody.CcData.Num, &wGroup, errChan)
	//

	// check credit card #
	go checkCreditCard(w, uuidStr, atoiCC, &wGroup, errChan)
	//

	// check expiry is in the future
	go checkExpiry(w, uuidStr, parsedBody.CcData.Expiry, &wGroup, errChan)
	//

	// mock cvv check, if sum equals 8 then fail (makes testing possible)
	go checkCvv(w, uuidStr, parsedBody.CcData.Cvv, &wGroup, errChan)
	//

	wGroup.Wait()
	// check if theres an error, continue otherwise
	select {
	case err = <-errChan:
		log.Println(err)
		return
	default:
	}

	//
	db, err = badger.Open(badger.DefaultOptions("./db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = submitToDB(w, db, &parsedBody, uuid, uuidStr)

}

// Add this transaction to the DB
func submitToDB(w http.ResponseWriter, db *badger.DB, parsedBody *InputData, uuid []byte, uuidStr string) error {

	// construct our data to be stored
	moneyData := MoneyData{
		Currency:           parsedBody.MoneyData.Currency,
		OutstandingBalance: parsedBody.MoneyData.Amount,
		Amount:             parsedBody.MoneyData.Amount,
	}

	storeStruct := &StoredData{Data: InputData{
		CcData:    parsedBody.CcData,
		MoneyData: moneyData,
	}, Voided: false,
		Refunded: false}

	marshalledBody, _ := json.Marshal(storeStruct)

	// store values
	err := db.Update(func(txn *badger.Txn) error {
		innerErr := txn.Set([]byte(uuid), []byte(marshalledBody))

		return innerErr
	})

	if err != nil {
		fmt.Fprint(w, err.Error())
		log.Println(err)
		return err
	}

	//return a json string to the caller
	retStr := fmt.Sprintf(`{
	"uid" : "%v",
	"success" : true,
	"amount" : "%v",
	"currency" : "%v"
}`, uuidStr, parsedBody.MoneyData.Amount, parsedBody.MoneyData.Currency)

	fmt.Fprintln(w, retStr)
	return nil
}

// Check for the specified edge case
func checkEdgeCaseAuth(w http.ResponseWriter, uuid string, num string, wg *sync.WaitGroup, errChan chan error) {
	if num == "4000000000000119" {
		errStr := fmt.Sprintln(`{ 
	"uid" : "$s"
	"success" : false,
	"message" : "Authorisation failure"
}`, uuid)
		fmt.Fprint(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("Authorisation Failure")
	}
	wg.Done()
}

// Check the credit card number passes a luhn check
func checkCreditCard(w http.ResponseWriter, uuid string, ccnum int, wg *sync.WaitGroup, errChan chan error) {
	if !luhn.Valid(ccnum) {
		errStr := fmt.Sprintf(`{ 
	"uid" : "%v"
	"success" : false,
	"message" : "Credit card number is not valid"
}`, uuid)
		fmt.Fprint(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("Credit card number is not valid")
	}
	wg.Done()
}

// Check the expiry date is in the future
func checkExpiry(w http.ResponseWriter, uuid string, expiry string, wg *sync.WaitGroup, errChan chan error) {
	now := time.Now()

	date := strings.Split(expiry, "/")
	month := date[0]
	year := "20" + date[1]

	dateStr := year + "-" + month + "-01"

	parsedExpiry, _ := time.Parse("2006-01-02", dateStr)

	if now.After(parsedExpiry) {
		errStr := fmt.Sprintln(`{ 
	"uid" : "$s"
	"success" : false,
	"message" : "Card Expired"
}`, uuid)
		fmt.Fprint(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("Card Expired")
	}
	wg.Done()
}

// "Check" the CVV number is "correct" (stubbed for testing purposes)
func checkCvv(w http.ResponseWriter, uuid string, cvv string, wg *sync.WaitGroup, errChan chan error) {
	sum := 0

	for i := 0; i < 3; i++ {
		val, _ := strconv.Atoi(string(cvv[i]))

		sum += val
	}

	if sum == 8 {
		errStr := fmt.Sprintln(`{ 
	"uid" : "$s"
	"success" : false,
	"message" : "CVV Incorrect"
}`, uuid)
		fmt.Fprint(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("CVV Incorrect")
	}
	wg.Done()
}
