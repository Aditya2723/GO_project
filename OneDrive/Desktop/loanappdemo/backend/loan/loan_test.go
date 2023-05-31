package loan

import (
	con "backend/Config"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestInsertLoan(t *testing.T) {

	config, err := con.LoadConfig("Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	// Create a loan object for testing
	newLoan := Loan_details{
		Loan_type:            "home",
		Loan_amount:          60000,
		Tenure:               4,
		Pincode:              123456,
		Employment_type:      "Salaried",
		Gross_monthly_income: 5000,
	}

	// Marshal the loan object to JSON
	body, err := json.Marshal(newLoan)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP request with the loan JSON payload
	req, err := http.NewRequest("POST", "/loans", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}

	// Set the request headers
	req.Header.Set("Content-type", "application/json")

	// Create a new HTTP response recorder
	rr := httptest.NewRecorder()

	// Create a new router
	r := mux.NewRouter()
	r.HandleFunc("/loans", InsertLoanDetails).Methods("POST")

	// Make the HTTP request and record the response
	r.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

}

func TestUpdateLoan(t *testing.T) {

	config, err := con.LoadConfig("Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	// Set up test data
	loanID := "12"
	updateLoan := Loan_details{Loan_type: "test", Loan_amount: 60000, Tenure: 4, Pincode: 123456, Employment_type: "Salaried", Gross_monthly_income: 40000}

	// Create request body
	body, err := json.Marshal(updateLoan)
	if err != nil {
		t.Fatal(err)
	}

	// Create HTTP PUT request with loan ID parameter
	req, err := http.NewRequest("PUT", "/update/"+loanID, strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}

	// Set request headers
	req.Header.Set("Content-type", "application/json")

	// Create router and HTTP response recorder
	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/update/{id}", UpdateLoanDetails).Methods("PUT")

	// Make HTTP PUT request and record response
	r.ServeHTTP(rr, req)

	// Check response status code
	if status := rr.Code; status < 200 || status > 299 {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check response body
	expected := "Loan ID updated successfully"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestDeleteLoan(t *testing.T) {

	config, err := con.LoadConfig("Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	// Set up test data
	loanID := "36"

	// Create HTTP DELETE request with loan ID parameter
	req, err := http.NewRequest("DELETE", "/loan/"+loanID, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create router and HTTP response recorder
	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/loan/{id}", DeleteLoanDetails).Methods("DELETE")

	// Make HTTP DELETE request and record response
	r.ServeHTTP(rr, req)

	// Check response status code
	if status := rr.Code; status < 200 || status > 299 {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
func TestLoanIndex(t *testing.T) {

	config, err := con.LoadConfig("Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	req, err := http.NewRequest("GET", "/get/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/get/{id}", GetLoanDetails).Methods("GET")
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"id":1,"loan_type":"test","loan_amount":60000,"tenure":4,"pincode":123456,"created_at":"0001-01-01T00:00:00Z","last_modified":"0001-01-01T00:00:00Z","employment_type":"Salaried","gross_monthly_income":40000}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
