package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	badger "github.com/dgraph-io/badger/v2"
)

/*
{
	"UID" : string
}
*/

// Void is the method assigned to /void
func Void(w http.ResponseWriter, r *http.Request) {

	var parsedBody VoidObj //marshal the json into a nice struct

	file, fileErr := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	defer file.Close()

	if fileErr != nil {
		log.SetOutput(file)
	}

	err := json.NewDecoder(r.Body).Decode(&parsedBody)

	if err != nil {
		log.Fatal(err)
	}
	db, err := badger.Open(badger.DefaultOptions("./db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	writeVoidToDB(w, db, &parsedBody)

}

// Write the void operation to disk
func writeVoidToDB(w http.ResponseWriter, db *badger.DB, parsedBody *VoidObj) error {

	var voidedVal StoredData

	err := db.Update(func(txn *badger.Txn) error {
		uid := []byte(parsedBody.UID)
		err := queryAndGet(parsedBody.UID, txn, &voidedVal)

		if err != nil {
			return err
		}

		if voidedVal.Voided {
			// already voided, so just return
			return nil
		}

		voidedVal.Voided = true
		marshalledVoidedVal, _ := json.Marshal(voidedVal)
		err = txn.Set(uid, marshalledVoidedVal)

		return err
	})

	if err != nil {
		fmt.Fprint(w, err.Error())
		log.Println(err)
		return err
	}

	retStr := fmt.Sprintf(`{
	"success" : true,
	"amount" : "%v",
	"currency" : "%v"
}`, voidedVal.Data.MoneyData.OutstandingBalance, voidedVal.Data.MoneyData.Currency)

	fmt.Fprintln(w, retStr)
	return nil
}
