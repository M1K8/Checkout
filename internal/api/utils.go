package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	badger "github.com/dgraph-io/badger/v2"
)

// Within the passed transaction, query a given key, and return the result.
func queryAndGet(key string, txn *badger.Txn, retObj interface{}) error {
	uid := []byte(key)
	item, err := txn.Get(uid)

	if err != nil {
		return err
	}

	err = item.Value(func(val []byte) error {
		err = json.Unmarshal(val, &retObj)
		return err
	})

	return nil
}

// Check if the transaction is voided
func checkVoid(w http.ResponseWriter, uuid string, captureVal *StoredData, wg *sync.WaitGroup, errChan chan error) {
	if captureVal.Voided {
		errStr := fmt.Sprintf(`{
	"success" : false,
	"amount" : "%v",
	"currency" : "%v",
	"error" : "The transaction is voided"
}`, captureVal.Data.MoneyData.OutstandingBalance, captureVal.Data.MoneyData.Currency)

		fmt.Fprintln(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("Transaction voided")
	}
	wg.Done()
}

// Check if the transaction is refunded
func checkRefunded(w http.ResponseWriter, uuid string, captureVal *StoredData, wg *sync.WaitGroup, errChan chan error) {
	if captureVal.Refunded {
		errStr := fmt.Sprintf(`{
	"success" : false,
	"amount" : "%v",
	"currency" : "%v",
	"error" : "Cannot capture - transaction has been refunded"
}`, captureVal.Data.MoneyData.OutstandingBalance, captureVal.Data.MoneyData.Currency)

		fmt.Fprintln(w, errStr)
		log.Println("TXN: " + uuid + " || " + errStr)
		errChan <- errors.New("Transaction refunded")
	}
	wg.Done()
}
