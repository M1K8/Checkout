package api

import (
	"log"
	"net/http/httptest"
	"sync"
	"testing"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/shopspring/decimal"
)

func TestCheckEdgeCaseRefundFalse(t *testing.T) {

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkEdgeCaseRefund(mockWriter, "test1", "4000000000003248", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() == ("Refund failure") {
			t.Error("Refund Edge case test failed: expected valid but got invalid")
		}
	default:

	}
}

func TestCheckEdgeCaseRefundTrue(t *testing.T) {

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkEdgeCaseRefund(mockWriter, "test1", "4000000000003238", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("Refund failure") {
			t.Error("Refund Edge case test failed: " + err.Error())
		}
	default:
		t.Error("Refund Edge case test failed: expected invalid but got valid")
	}
}

func TestCheckRefundBalanceFalse(t *testing.T) {
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "5", OutstandingBalance: "5"}
	mockCcData := CcData{Num: "7070707070707070", Expiry: "01/22", Cvv: "021"}
	mockData := &StoredData{Data: InputData{CcData: mockCcData, MoneyData: mockMoneyData}, Voided: false}

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)
	mockIncoming, _ := decimal.NewFromString("3")
	mockCurrent, _ := decimal.NewFromString("2")
	mockTotal, _ := decimal.NewFromString("5")

	checkBalanceRefund(mockWriter, "test1", mockIncoming, mockTotal, mockCurrent, mockData, &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("TheThe amount specified in the request is greater than the current balance") {
			t.Error("Current balance case test failed: " + err.Error())
		}
	default:

	}
}

func TestCheckRefundBalanceTrue(t *testing.T) {
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "5", OutstandingBalance: "5"}
	mockCcData := CcData{Num: "818181818181818", Expiry: "01/22", Cvv: "021"}
	mockData := &StoredData{Data: InputData{CcData: mockCcData, MoneyData: mockMoneyData}, Voided: false}

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)
	mockIncoming, _ := decimal.NewFromString("3")
	mockCurrent, _ := decimal.NewFromString("4")
	mockTotal, _ := decimal.NewFromString("5")

	checkBalanceRefund(mockWriter, "test2", mockIncoming, mockTotal, mockCurrent, mockData, &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("The amount specified in the request is greater than the current balance") {
			t.Error("Current balance case test failed: " + err.Error())
		}
	default:
		t.Error("Current balance case test failed: expected true but got false")

	}
}

func TestWriteRefundToDB(t *testing.T) {
	mockWriter := httptest.NewRecorder()
	db, err := badger.Open(badger.DefaultOptions("./testdb"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "50", OutstandingBalance: "50"}
	mockCcData := CcData{Num: "4343434343434334", Expiry: "01/22", Cvv: "021"}

	err = submitToDB(mockWriter, db, &InputData{CcData: mockCcData, MoneyData: mockMoneyData}, []byte("TestRefund"), "TestRefund")

	if err != nil {
		t.Error(err.Error())
	}

	var expectedObj StoredData

	db.View(func(txn *badger.Txn) error {
		queryAndGet("TestRefund", txn, &expectedObj)

		if expectedObj.Data.CcData.Num != "4343434343434334" {
			t.Error("Error - Data not correctly saved, ccnum was saved as " + expectedObj.Data.CcData.Num)
		}
		return nil
	})

	mockCurr, _ := decimal.NewFromString("25")
	mockInc, _ := decimal.NewFromString("25")

	writeRefundToDB(mockWriter, db, []byte("TestRefund"), mockCurr, mockInc, &expectedObj)

	db.View(func(txn *badger.Txn) error {
		queryAndGet("TestRefund", txn, &expectedObj)

		if expectedObj.Data.MoneyData.OutstandingBalance != "50" {
			t.Error("Error - Refund failure " + expectedObj.Data.MoneyData.OutstandingBalance)
		}

		if !expectedObj.Refunded {
			t.Error("Error - Refund failure: transaction not refunded")
		}
		return nil
	})
}
