package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	badger "github.com/dgraph-io/badger/v2"

	"github.com/shopspring/decimal"
)

// Capture is the method assigned to /capture
func Capture(w http.ResponseWriter, r *http.Request) {

	var currentBalance decimal.Decimal
	var incomingAmount decimal.Decimal
	var captureVal StoredData

	file, fileErr := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if fileErr == nil {
		log.SetOutput(file)
		defer file.Close()
	}

	db, err := badger.Open(badger.DefaultOptions("./db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var parsedBody TransactionObj //marshal the json into a nice struct
	wGroup := sync.WaitGroup{}
	wGroup.Add(4)
	errChan := make(chan error, 4)

	err = json.NewDecoder(r.Body).Decode(&parsedBody)

	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(txn *badger.Txn) error {
		incomingAmount, _ = decimal.NewFromString(parsedBody.Amount)

		err = queryAndGet(parsedBody.UID, txn, &captureVal)

		if err != nil {
			fmt.Fprintln(w, err.Error())
			log.Fatal(err)
			return err
		}

		// 'edge' case
		go checkEdgeCaseCapture(w, parsedBody.UID, captureVal.Data.CcData.Num, &wGroup, errChan)
		//

		// check void
		go checkVoid(w, parsedBody.UID, &captureVal, &wGroup, errChan)
		//

		// check refunded
		go checkRefunded(w, parsedBody.UID, &captureVal, &wGroup, errChan)
		//

		// check balance
		currentBalance, _ = decimal.NewFromString(captureVal.Data.MoneyData.OutstandingBalance)
		go checkBalanceCapture(w, parsedBody.UID, incomingAmount, currentBalance, &captureVal, &wGroup, errChan)
		//

		wGroup.Wait()
		// check if theres an error, continue otherwise
		select {
		case err = <-errChan:
			fmt.Println(err)
			log.Println("TXN: " + parsedBody.UID + " || " + err.Error())
			return err
		default:
		}
		return nil
	})

	if err == nil {
		writeCaptureToDB(w, db, []byte(parsedBody.UID), currentBalance, incomingAmount, &captureVal)
	} else {
		log.Println(err)
	}

}

// Write the captured amount to disk
func writeCaptureToDB(w http.ResponseWriter, db *badger.DB, uuid []byte, currentBalance decimal.Decimal, incomingAmount decimal.Decimal, captureVal *StoredData) error {
	fmt.Println(1)
	err := db.Update(func(txn *badger.Txn) error {
		captureVal.Data.MoneyData.OutstandingBalance = currentBalance.Sub(incomingAmount).String()
		marshalledCaptureVal, _ := json.Marshal(captureVal)
		err := txn.Set(uuid, marshalledCaptureVal)
		if err != nil {
			fmt.Fprint(w, err.Error())
			log.Println("TXN: " + string(uuid) + " || " + err.Error())
			return err
		}

		retStr := fmt.Sprintf(`{
	"success" : true,
	"amount" : "%v",
	"currency" : "%v"
}`, captureVal.Data.MoneyData.OutstandingBalance, captureVal.Data.MoneyData.Currency)

		fmt.Fprintln(w, retStr)
		return nil
	})

	return err
}

// Check for the specified edge case
func checkEdgeCaseCapture(w http.ResponseWriter, uuid string, num string, wg *sync.WaitGroup, errChan chan error) {
	if num == "4000000000000259" {
		errStr := fmt.Sprintln(`{ 
	"uid" : "$s"
	"success" : false,
	"message" : "Capture failure"
}`, uuid)
		fmt.Fprint(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("Capture Failure")
	}
	wg.Done()
}

// Check the incoming amount is not greater than the outstanding balance
func checkBalanceCapture(w http.ResponseWriter, uuid string, incomingAmount decimal.Decimal, currentBalance decimal.Decimal, captureVal *StoredData, wg *sync.WaitGroup, errChan chan error) {
	if (incomingAmount.GreaterThan(currentBalance)) || incomingAmount.IsNegative() {
		errStr := fmt.Sprintf(`{
	"success" : false,
	"amount" : "%v",
	"currency" : "%v",
	"error" : "The amount specified in the request is greater than the remaining balance"
}`, captureVal.Data.MoneyData.OutstandingBalance, captureVal.Data.MoneyData.Currency)

		fmt.Fprintln(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("The amount specified in the request is greater than the remaining balance")
	}
	wg.Done()
}
