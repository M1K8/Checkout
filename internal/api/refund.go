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

// Refund is the method assigned to /refund
func Refund(w http.ResponseWriter, r *http.Request) {
	var refundVal StoredData
	var totalAmount decimal.Decimal
	var incomingAmount decimal.Decimal
	var currentBalance decimal.Decimal

	file, fileErr := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	defer file.Close()

	if fileErr != nil {
		log.SetOutput(file)
	}

	db, err := badger.Open(badger.DefaultOptions("./db"))
	wGroup := sync.WaitGroup{}
	wGroup.Add(3)
	errChan := make(chan error, 3)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	var parsedBody TransactionObj //marshal the json into a nice struct
	err = json.NewDecoder(r.Body).Decode(&parsedBody)
	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(txn *badger.Txn) error {
		incomingAmount, _ = decimal.NewFromString(parsedBody.Amount)

		err = queryAndGet(parsedBody.UID, txn, &refundVal)

		if err != nil {
			fmt.Fprintln(w, err.Error())
			return err
		}

		// 'edge' case
		go checkEdgeCaseRefund(w, parsedBody.UID, refundVal.Data.CcData.Num, &wGroup, errChan)
		//

		// check void
		go checkVoid(w, parsedBody.UID, &refundVal, &wGroup, errChan)
		//

		// check balance
		currentBalance, _ = decimal.NewFromString(refundVal.Data.MoneyData.OutstandingBalance)
		totalAmount, _ = decimal.NewFromString(refundVal.Data.MoneyData.Amount)
		go checkBalanceRefund(w, parsedBody.UID, incomingAmount, totalAmount, currentBalance, &refundVal, &wGroup, errChan)
		//

		wGroup.Wait()
		// check if theres an error, continue otherwise
		select {
		case err = <-errChan:
			log.Println(err)
			return err
		default:
			return nil
		}
	})
	if err == nil {
		writeRefundToDB(w, db, []byte(parsedBody.UID), currentBalance, incomingAmount, &refundVal)
	} else {
		log.Println(err)
	}

}

// Check for the specified edge case
func checkEdgeCaseRefund(w http.ResponseWriter, uuid string, num string, wg *sync.WaitGroup, errChan chan error) {
	if num == "4000000000003238" {
		errStr := fmt.Sprintln(`{ 
"uid" : "$s"
"success" : false,
"message" : "Refund failure"
}`, uuid)

		fmt.Fprint(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("Refund failure")
	}
	wg.Done()
}

// Check the incoming amount is not greater than the (total amount - curent balance)
func checkBalanceRefund(w http.ResponseWriter, uuid string, incomingAmount decimal.Decimal, totalAmount decimal.Decimal, currentBalance decimal.Decimal, refundVal *StoredData, wg *sync.WaitGroup, errChan chan error) {
	if (incomingAmount.GreaterThan(totalAmount.Sub(currentBalance))) || incomingAmount.IsNegative() {
		errStr := fmt.Sprintf(`{
	"success" : false,
	"amount" : "%v",
	"currency" : "%v",
	"error" : "The amount specified in the request is greater than the current balance"
}`, refundVal.Data.MoneyData.OutstandingBalance, refundVal.Data.MoneyData.Currency)

		fmt.Fprintln(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("The amount specified in the request is greater than the current balance")
	}
	wg.Done()
}

// Write the refund to disk
func writeRefundToDB(w http.ResponseWriter, db *badger.DB, uuid []byte, currentBalance decimal.Decimal, incomingAmount decimal.Decimal, refundVal *StoredData) error {
	err := db.Update(func(txn *badger.Txn) error {

		refundVal.Data.MoneyData.OutstandingBalance = currentBalance.Add(incomingAmount).String()

		total, _ := (decimal.NewFromString(refundVal.Data.MoneyData.Amount))
		if currentBalance.Add(incomingAmount).Equal(total) {
			refundVal.Voided = true
		}

		refundVal.Refunded = true

		marshalledRefundVal, _ := json.Marshal(refundVal)

		err := txn.Set([]byte(uuid), marshalledRefundVal)

		if err != nil {
			fmt.Fprint(w, err.Error())
			log.Println(err)
			return err
		}

		retStr := fmt.Sprintf(`{
	"success" : true,
	"amount" : "%v",
	"currency" : "%v"
}`, refundVal.Data.MoneyData.OutstandingBalance, refundVal.Data.MoneyData.Currency)

		fmt.Fprintln(w, retStr)
		return nil
	})
	return err
}
