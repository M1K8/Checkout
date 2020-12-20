package api

import (
	"log"
	"net/http/httptest"
	"sync"
	"testing"

	badger "github.com/dgraph-io/badger/v2"
)

func TestCheckEdgeCaseAuthTrue(t *testing.T) {

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkEdgeCaseAuth(mockWriter, "test1", "4000000000000119", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("Authorisation Failure") {
			t.Error("Edge case test failed:" + err.Error())
		}
	default:
		t.Error("Edge case test failed: expected invalid but got valid")
	}
}

func TestCheckEdgeCaseAuthFalse(t *testing.T) {

	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkEdgeCaseAuth(mockWriter, "test1", "4000000000000118", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() == ("Authorisation Failure") {
			t.Error("Edge case test failed: expected valid but got invalid")
		}
	default:

	}
}

func TestCheckCreditCardInvalid(t *testing.T) {
	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkCreditCard(mockWriter, "test1", 123456, &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("Credit card number is not valid") {
			t.Error("Creadit card validity test failed")
		}
	default:
		t.Error("Credit card validity test failed: expected invalid but got valid")
	}
}

func TestCheckCreditCardValid(t *testing.T) {
	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkCreditCard(mockWriter, "test1", 4024007156809708, &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("Credit card number is not valid") {
			t.Error("Credit card validity test failed: expected valid but got invalid")
		}
	default:
	}
}

func TestCheckExpiryValid(t *testing.T) {
	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkExpiry(mockWriter, "test1", "01/22", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("Card Expired") {
			t.Error("Credit card expiry test failed: expected valid but got invalid")
		}
	default:
	}
}

func TestCheckExpiryInvalid(t *testing.T) {
	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkExpiry(mockWriter, "test1", "01/02", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("Card Expired") {
			t.Error("Credit card expiry test failed")
		}
	default:
		t.Error("Credit card expiry test failed: expected invalid but got valid")
	}
}

func TestCheckCvvInvalid(t *testing.T) {
	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkCvv(mockWriter, "test1", "800", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() != ("CVV Incorrect") {
			t.Error("Credit card CVV test failed")
		}
	default:
		t.Error("Credit card CVV test failed: expected invalid but got valid")
	}
}

func TestCheckCvvValid(t *testing.T) {
	mockWriter := httptest.NewRecorder()
	wGroup := sync.WaitGroup{}
	wGroup.Add(1)
	errChan := make(chan error, 1)

	checkCvv(mockWriter, "test1", "840", &wGroup, errChan)

	wGroup.Wait()

	select {
	case err := <-errChan:
		if err.Error() == ("CVV Incorrect") {
			t.Error("Credit card CVV test failed: expected valid but got invalid")
		}
	default:

	}
}

func TestDBWrite(t *testing.T) {
	mockWriter := httptest.NewRecorder()
	db, err := badger.Open(badger.DefaultOptions("./testdb"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mockMoneyData := MoneyData{Currency: "GBP", Amount: "50", OutstandingBalance: "50"}
	mockCcData := CcData{Num: "5050505050505050", Expiry: "01/22", Cvv: "021"}

	err = submitToDB(mockWriter, db, &InputData{CcData: mockCcData, MoneyData: mockMoneyData}, []byte("test123"), "test123")

	if err != nil {
		t.Error(err.Error())
	}

	var expectedObj StoredData

	db.View(func(txn *badger.Txn) error {
		queryAndGet("test123", txn, &expectedObj)

		if expectedObj.Data.CcData.Num != "5050505050505050" {
			t.Error("Error - Data not correctly saved, ccnum was saved as " + expectedObj.Data.CcData.Num)
		}
		return nil
	})
}
