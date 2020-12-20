package api

// Struct containing a request to /void
type VoidObj struct {
	UID string `json:"UID"`
}

// Struct containing a request to /capture and /refund
type TransactionObj struct {
	UID    string `json:"UID"`
	Amount string `json:"Amount"`
}

// Struct containing the Credit Card data
type CcData struct {
	Num    string `json:"Num"`
	Expiry string `json:"Expiry"`
	Cvv    string `json:"Cvv"`
}

// Struct containing data relating to money
type MoneyData struct {
	Currency           string `json:"Currency"`
	Amount             string `json:"Amount"`
	OutstandingBalance string `json:"OutstandingBalance"`
}

// Struct containing the key data to be marhsalled and stored in the DB
type InputData struct {
	CcData    CcData    `json:"Ccdata"`
	MoneyData MoneyData `json:"MoneyData"`
}

// Struct representing the JSOn object written to the DB
type StoredData struct {
	Data     InputData `json:"Data"`
	Voided   bool      `json:"Voided"`
	Refunded bool      `json:"Refunded"`
}
