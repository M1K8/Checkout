package api

import (
	"log"
	"net/http/httptest"
	"testing"

	badger "github.com/dgraph-io/badger/v2"
)

func TestWriteVoid(t *testing.T) {
	// write our value to DB:
	mockWriter := httptest.NewRecorder()
	db, err := badger.Open(badger.DefaultOptions("./testdb"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "500", OutstandingBalance: "500"}
	mockCcData := CcData{Num: "4040404040404040", Expiry: "01/33", Cvv: "021"}

	err = submitToDB(mockWriter, db, &InputData{CcData: mockCcData, MoneyData: mockMoneyData}, []byte("VoidTest"), "VoidTest")

	if err != nil {
		t.Error(err.Error())
	}

	var expectedObj StoredData

	db.View(func(txn *badger.Txn) error {
		queryAndGet("VoidTest", txn, &expectedObj)

		if expectedObj.Data.CcData.Num != "4040404040404040" {
			t.Error("Error - Data not correctly saved, ccnum was saved as " + expectedObj.Data.CcData.Num)
		}
		return nil
	})
	//

	err = writeVoidToDB(mockWriter, db, &VoidObj{UID: "VoidTest"})

	if err != nil {
		log.Fatal(err)
	}

	db.View(func(txn *badger.Txn) error {
		queryAndGet("VoidTest", txn, &expectedObj)

		if !expectedObj.Voided {
			t.Error("Error - Void was not applied")
		}
		return nil
	})
}
