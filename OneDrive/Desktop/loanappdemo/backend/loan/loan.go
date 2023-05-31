package loan

import (
	con "backend/Config"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	_ "github.com/go-sql-driver/mysql"

	"net/http"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"

	"github.com/gorilla/mux"
)

type Loan_details struct {
	ID                   int       `json:"id,omitempty"`
	Loan_type            string    `json:"loan_type,omitempty" validate:"required"`
	Loan_amount          float64   `json:"loan_amount,omitempty" validate:"gt=0"`
	Tenure               int       `json:"tenure,omitempty" validate:"gt=0"`
	Pincode              int       `json:"pincode,omitempty" validate:"len=6"`
	Created_at           time.Time `json:"created_at,omitempty"`
	Last_modified        time.Time `json:"last_modified,omitempty"`
	Employment_type      string    `json:"employment_type,omitempty" `
	Gross_monthly_income float64   `json:"gross_monthly_income"`
	Status               string    `json:"status"`
}

var db *sql.DB
var loan Loan_details

func init() {
	var err error
	db, err = con.GetDB()
	if err != nil {
		panic(fmt.Errorf("failed to initialize database: %v", err))
	}
}

func ValidateLoan(loan Loan_details, validate *validator.Validate) error {
	// Check that loan_type is not empty
	if loan.Loan_type == "" {
		return errors.New("Loan type is required.")
	}

	// Check that loan_amount is greater than zero
	if loan.Loan_amount <= 0 {
		return errors.New("Loan amount must be greater than zero.")
	}

	// Check that tenure is greater than zero
	if loan.Tenure <= 0 {
		return errors.New("Tenure must be greater than zero.")
	}

	// Check that pincode is exactly 6 digits
	if match, _ := regexp.MatchString(`^\d{6}$`, fmt.Sprint(loan.Pincode)); !match {
		return errors.New("Pincode must be exactly 6 digits.")
	}

	// Check that employment_type is not empty
	if loan.Employment_type == "" {
		return errors.New("Employment type is required.")
	}

	if loan.Gross_monthly_income <= 0 {
		return errors.New("Gross monthly income must be greater than zero.")
	}

	// If all checks pass, return nil
	return nil
}

type Error struct {
	Message string `json:"message"`
}
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error *Error      `json:"error,omitempty"`
}

func InsertLoanDetails(w http.ResponseWriter, r *http.Request) {

	config, err := con.LoadConfig("Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	var loan Loan_details
	if err := json.NewDecoder(r.Body).Decode(&loan); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&Response{Error: &Error{Message: err.Error()}})
		return
	}

	// create validator instance
	validate := validator.New()

	// validate loan details
	if err := ValidateLoan(loan, validate); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&Response{Error: &Error{Message: err.Error()}})
		return
	}

	result, err := db.Exec("INSERT INTO loan_details_table(loan_type,loan_amount,pincode,tenure,employment_type,gross_monthly_income,credit_score,created_at,last_modified) VALUES(?,?,?,?,?,?,NOW(),NOW())", loan.Loan_type, loan.Loan_amount, loan.Pincode, loan.Tenure, loan.Employment_type, loan.Gross_monthly_income)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Error: &Error{Message: err.Error()}})
		return
	}

	id, _ := result.LastInsertId()
	loan.ID = int(id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&Response{Data: loan})
}

func UpdateLoanDetails(w http.ResponseWriter, r *http.Request) {

	config, err := con.LoadConfig("Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	params := mux.Vars(r)
	LoanID := params["id"]

	var loan Loan_details
	fmt.Println("r.body", r.Body)
	error := json.NewDecoder(r.Body).Decode(&loan)
	if error != nil {
		http.Error(w, error.Error(), http.StatusBadRequest)
		return
	}
	result, err := db.Exec("UPDATE loan_details_table SET loan_type=?, loan_amount=?, tenure=?, Pincode=?, employment_type=?, gross_monthly_income=?  WHERE id=?", loan.Loan_type, loan.Loan_amount, loan.Tenure, loan.Pincode, loan.Employment_type, loan.Gross_monthly_income, LoanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Loan ID not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Loan ID updated successfully")

	// err = templates.ExecuteTemplate(w, "http://localhost:9000/form/update_loan.html", loan)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

}

func DeleteLoanDetails(w http.ResponseWriter, r *http.Request) {

	config, err := con.LoadConfig("Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	params := mux.Vars(r)
	loanID := params["id"]

	result, err := db.Exec("DELETE FROM loan_details_table WHERE id=?", loanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Loan not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "loan deleted successfully")
}

var templates = template.Must(template.ParseFiles("form/loan.html"))

func GetLoanDetails(w http.ResponseWriter, r *http.Request) {
	config, err := con.LoadConfig("./Config/config.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	params := mux.Vars(r)
	loanID := params["id"]

	rows, err := db.Query("SELECT id, loan_type, loan_amount, tenure, pincode, employment_type, gross_monthly_income, status, created_at, last_modified FROM loan_details_table WHERE id=?", loanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()
	var createdAtStr string
	var lastAtStr string
	var loans []Loan_details

	for rows.Next() {
		var loan Loan_details
		//var statusText string
		err = rows.Scan(&loan.ID, &loan.Loan_type, &loan.Loan_amount, &loan.Tenure, &loan.Pincode, &loan.Employment_type, &loan.Gross_monthly_income, &loan.Status, &createdAtStr, &lastAtStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lastAt, err := time.Parse("2006-01-02 15:04:05", lastAtStr)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		loan.Created_at = createdAt
		loan.Last_modified = lastAt

		loans = append(loans, loan)
	}

	response := Response{
		Data: loans,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

//retrieve all data from loan

func LoanIndex(w http.ResponseWriter, r *http.Request) {
	config, err := con.LoadConfig("./Config/config.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	rows, err := db.Query("SELECT id, loan_type, loan_amount, tenure, pincode, employment_type, gross_monthly_income, status FROM loan_details_table")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var loans []Loan_details

	for rows.Next() {
		var loan Loan_details

		err = rows.Scan(&loan.ID, &loan.Loan_type, &loan.Loan_amount, &loan.Tenure, &loan.Pincode, &loan.Employment_type, &loan.Gross_monthly_income, &loan.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//loan.Status = getStatusValue(statusText)
		loans = append(loans, loan)
	}

	response := Response{
		Data: loans,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}
