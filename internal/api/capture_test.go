package api

import (
	"log"
	"net/http/httptest"
	"sync"
	"testing"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/shopspring/decimal"
)

func TestCheckEdgeCaseCaptureCatchesCorrectError(t *testing.T) {

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkEdgeCaseCapture(mockWriter, "test1", "4000000000000259", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("Capture Failure") {
			t.Error("Edge case test failed: " + err.Error())
		}
	default:
		t.Error("Edge case test failed: expected invalid but got valid")
	}
}

func TestCheckEdgeCaseCaptureFalse(t *testing.T) {

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkEdgeCaseCapture(mockWriter, "test11", "4000000000000118", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() == ("Capture failure") {
			t.Error("Capture Edge case test failed: expected valid but got invalid")
		}
	default:

	}
}

func TestCheckEdgeCaseCaptureTrue(t *testing.T) {

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkEdgeCaseCapture(mockWriter, "test12", "4000000000000259", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("Capture Failure") {
			t.Error("Capture Edge case test failed: " + err.Error())
		}
	default:
		t.Error("Capture Edge case test failed: expected invalid but got valid")

	}
}

func TestCheckVoidTrue(t *testing.T) {
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "50", OutstandingBalance: "50"}
	mockCcData := CcData{Num: "5050505050505050", Expiry: "01/22", Cvv: "021"}
	mockData := &StoredData{Data: InputData{CcData: mockCcData, MoneyData: mockMoneyData}, Voided: true}

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkVoid(mockWriter, "test", mockData, &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("Transaction voided") {
			t.Error("Void case test failed: " + err.Error())
		}
	default:
		t.Error("Void case test failed: expected true but got false")
	}
}

func TestCheckVoidFalse(t *testing.T) {
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "50", OutstandingBalance: "50"}
	mockCcData := CcData{Num: "5050505050505050", Expiry: "01/22", Cvv: "021"}
	mockData := &StoredData{Data: InputData{CcData: mockCcData, MoneyData: mockMoneyData}, Voided: false}

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkVoid(mockWriter, "test", mockData, &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() == ("Transaction voided") {
			t.Error("Void case test failed: expected false but got true")
		}
	default:

	}
}

func TestCheckRefunded(t *testing.T) {
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "50", OutstandingBalance: "50"}
	mockCcData := CcData{Num: "5050505050505050", Expiry: "01/22", Cvv: "021"}
	mockData := &StoredData{Data: InputData{CcData: mockCcData, MoneyData: mockMoneyData}, Voided: false, Refunded: true}

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkVoid(mockWriter, "test", mockData, &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() == ("Transaction refunded") {
			t.Error("Void case test failed: expected false but got true")
		}
	default:

	}
}

func TestCheckCaptureBalanceFalse(t *testing.T) {
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "5", OutstandingBalance: "5"}
	mockCcData := CcData{Num: "6060606060606060", Expiry: "01/22", Cvv: "021"}
	mockData := &StoredData{Data: InputData{CcData: mockCcData, MoneyData: mockMoneyData}, Voided: false}

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)
	mockIncoming, _ := decimal.NewFromString("4")
	mockCurrent, _ := decimal.NewFromString("5")

	checkBalanceCapture(mockWriter, "test1", mockIncoming, mockCurrent, mockData, &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("The amount specified in the request is greater than the remaining balance") {
			t.Error("Current balance case test failed")
		}
	default:

	}
}

func TestCheckCaptureBalanceTrue(t *testing.T) {
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "5", OutstandingBalance: "5"}
	mockCcData := CcData{Num: "6060606060606060", Expiry: "01/22", Cvv: "021"}
	mockData := &StoredData{Data: InputData{CcData: mockCcData, MoneyData: mockMoneyData}, Voided: false}

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)
	mockIncoming, _ := decimal.NewFromString("5")
	mockCurrent, _ := decimal.NewFromString("4")

	checkBalanceCapture(mockWriter, "test2", mockIncoming, mockCurrent, mockData, &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("The amount specified in the request is greater than the remaining balance") {
			t.Error("Current balance case test failed")
		}
	default:
		t.Error("Current balance case test failed: expected true but got false")

	}
}

func TestWriteCaptureToDB(t *testing.T) {
	mockWriter := httptest.NewRecorder()
	db, err := badger.Open(badger.DefaultOptions("./testdb"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "50", OutstandingBalance: "50"}
	mockCcData := CcData{Num: "2020202020202020", Expiry: "01/22", Cvv: "021"}

	err = submitToDB(mockWriter, db, &InputData{CcData: mockCcData, MoneyData: mockMoneyData}, []byte("TestCapture"), "TestCapture")

	if err != nil {
		t.Error(err.Error())
	}

	var expectedObj StoredData

	db.View(func(txn *badger.Txn) error {
		queryAndGet("TestCapture", txn, &expectedObj)

		if expectedObj.Data.CcData.Num != "2020202020202020" {
			t.Error("Error - Data not correctly saved, ccnum was saved as " + expectedObj.Data.CcData.Num)
		}
		return nil
	})
	mockCurr, _ := decimal.NewFromString("50")
	mockInc, _ := decimal.NewFromString("49")

	writeCaptureToDB(mockWriter, db, []byte("TestCapture"), mockCurr, mockInc, &expectedObj)

	db.View(func(txn *badger.Txn) error {
		queryAndGet("TestCapture", txn, &expectedObj)

		if expectedObj.Data.MoneyData.OutstandingBalance != "1" {
			t.Error("Error - Capture failure " + expectedObj.Data.MoneyData.OutstandingBalance)
		}
		return nil
	})
}
