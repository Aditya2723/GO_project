package lead

import (
	con "backend/Config"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	//"strings"
	//"text/template"

	"log"
	"regexp"

	"github.com/go-playground/validator"
	_ "github.com/go-sql-driver/mysql"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"

	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Lead_info struct {
	ID                   int       `json:"id,omitempty"`
	Loan_type            string    `json:"loan_type,omitempty"`
	Loan_amount          float64   `json:"loan_amount,omitempty"`
	Tenure               int       `json:"tenure,omitempty"`
	Pincode              int       `json:"pincode,omitempty"`
	Employment_type      string    `json:"employment_type,omitempty" `
	Gross_monthly_income float64   `json:"gross_monthly_income"`
	Status               int       `json:"status,omitempty"`
	Created_at           time.Time `json:"created_at,omitempty"`
	Last_modified        time.Time `json:"last_modified,omitempty"`
}

var db *sql.DB

// var tmpl = template.Must(template.ParseGlob("form/*.html"))

func init() {
	var err error
	db, err = con.GetDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
}

func SetDB(database *sql.DB) {
	db = database
}

func ValidateLead(lead Lead_info, validate *validator.Validate) error {
	// Check that loan_type is not empty
	if lead.Loan_type == "" {
		return errors.New("Loan type is required.")
	}

	// Check that loan_amount is greater than zero
	if lead.Loan_amount <= 0 {
		return errors.New("Loan amount must be greater than zero.")
	}

	// Check that tenure is greater than zero
	if lead.Tenure <= 0 {
		return errors.New("Tenure must be greater than zero.")
	}

	// Check that pincode is exactly 6 digits
	if match, _ := regexp.MatchString(`^\d{6}$`, fmt.Sprint(lead.Pincode)); !match {
		return errors.New("Pincode must be exactly 6 digits.")
	}

	// Check that employment_type is not empty
	if lead.Employment_type == "" {
		return errors.New("Employment type is required.")
	}

	if lead.Gross_monthly_income <= 0 {
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

func InsertLead(w http.ResponseWriter, r *http.Request) {

	config, err := con.LoadConfig("./Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	var lead Lead_info
	if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&Response{Error: &Error{Message: err.Error()}})
		return
	}

	validate := validator.New()

	// validate loan details
	if err := ValidateLead(lead, validate); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&Response{Error: &Error{Message: err.Error()}})
		return
	}

	if lead.Gross_monthly_income > 20000 {
		lead.Status = 2
	} else {
		lead.Status = 3
	}

	var statusText string

	switch lead.Status {
	case 1:
		statusText = "Pending"
	case 2:
		statusText = "Approved"
	case 3:
		statusText = "Declined"
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&Response{Error: &Error{Message: "Invalid status"}})
		return
	}

	result, err := db.Exec("INSERT INTO lead_table (created_at, last_modified, status, loan_type, loan_amount, tenure, pincode, employment_type, gross_monthly_income) VALUES (NOW(), NOW(), ?,?,?,?,?,?,?)", statusText, lead.Loan_type, lead.Loan_amount, lead.Tenure, lead.Pincode, lead.Employment_type, lead.Gross_monthly_income)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Error: &Error{Message: err.Error()}})
		return
	}

	id, _ := result.LastInsertId()
	lead.ID = int(id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&Response{Data: lead})
}

func DeleteLead(w http.ResponseWriter, r *http.Request) {

	config, err := con.LoadConfig("./Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	params := mux.Vars(r)
	leadID := params["id"]

	result, err := db.Exec("DELETE FROM lead_table WHERE id=?", leadID)
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
		http.Error(w, "Lead not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Lead deleted successfully")
}

// Lead Dashboard code
func LeadIndexAll(w http.ResponseWriter, r *http.Request) {
	config, err := con.LoadConfig("./Config/config.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	params := mux.Vars(r)
	leadID := params["id"]

	rows, err := db.Query("SELECT id, loan_type, loan_amount, tenure, pincode, employment_type, gross_monthly_income, status FROM lead_table WHERE id=?", leadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var leads []Lead_info

	for rows.Next() {
		var lead Lead_info
		var statusText string
		err = rows.Scan(&lead.ID, &lead.Loan_type, &lead.Loan_amount, &lead.Tenure, &lead.Pincode, &lead.Employment_type, &lead.Gross_monthly_income, &statusText)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lead.Status = getStatusValue(statusText)
		leads = append(leads, lead)
	}

	response := Response{
		Data: leads,
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

func getStatusValue(statusText string) int {
	switch statusText {
	case "Pending":
		return 1
	case "Approved":
		return 2
	case "Declined":
		return 3
	default:
		return 0
	}
}

// var templates = template.Must(template.ParseFiles("form/update.html"))

func UpdateLead(w http.ResponseWriter, r *http.Request) {
	config, err := con.LoadConfig("Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	params := mux.Vars(r)
	LeadID := params["id"]

	// Check if the request method is allowed
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", http.MethodPatch)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var lead Lead_info
	//fmt.Println("r.Body ================== >", r.Body)
	err = json.NewDecoder(r.Body).Decode(&lead)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//fmt.Printf("%#v", lead)
	result, err := db.Exec("UPDATE lead_table SET loan_type=?, loan_amount=?, tenure=?, pincode=?, employment_type=?, gross_monthly_income=? WHERE id=?", lead.Loan_type, lead.Loan_amount, lead.Tenure, lead.Pincode, lead.Employment_type, lead.Gross_monthly_income, LeadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//fmt.Println(rowsAffected)

	if rowsAffected == 0 {
		http.Error(w, "Lead ID not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Lead ID updated successfully")

}

// retrieve all leads
func LeadIndex(w http.ResponseWriter, r *http.Request) {
	config, err := con.LoadConfig("./Config/config.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	rows, err := db.Query("SELECT id, loan_type, loan_amount, tenure, pincode, employment_type, gross_monthly_income, status FROM lead_table")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var leads []Lead_info

	for rows.Next() {
		var lead Lead_info
		var statusText string
		err = rows.Scan(&lead.ID, &lead.Loan_type, &lead.Loan_amount, &lead.Tenure, &lead.Pincode, &lead.Employment_type, &lead.Gross_monthly_income, &statusText)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lead.Status = getStatusValue(statusText)
		leads = append(leads, lead)
	}

	response := Response{
		Data: leads,
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

//trial

// func HandleUpdateLead(w http.ResponseWriter, r *http.Request) {
// 	config, err := con.LoadConfig("./Config/config.yaml")
// 	if err != nil {
// 		panic(err)
// 	}

// 	con.ConnectDB(config)
// 	defer con.CloseDB()

// 	params := mux.Vars(r)
// 	leadID := params["id"]

// 	var lead Lead_info
// 	err = json.NewDecoder(r.Body).Decode(&lead)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// Fetch the existing lead details from the database
// 	var existingLead Lead_info
// 	err = db.QueryRow("SELECT * FROM lead_table WHERE id = ?", leadID).Scan(
// 		&existingLead.ID,
// 		&existingLead.Loan_type,
// 		&existingLead.Loan_amount,
// 		&existingLead.Tenure,
// 		&existingLead.Pincode,
// 		&existingLead.Employment_type,
// 		&existingLead.Gross_monthly_income,
// 		&existingLead.Status,
// 		&existingLead.Created_at,
// 		&existingLead.Last_modified,
// 	)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Update the lead details with the submitted form data
// 	existingLead.Loan_type = lead.Loan_type
// 	existingLead.Loan_amount = lead.Loan_amount
// 	existingLead.Tenure = lead.Tenure
// 	existingLead.Pincode = lead.Pincode
// 	existingLead.Employment_type = lead.Employment_type
// 	existingLead.Gross_monthly_income = lead.Gross_monthly_income
// 	//existingLead.Status = lead.Status

// 	// Perform the database update
// 	_, err = db.Exec("UPDATE lead_table SET loan_type=?, loan_amount=?, tenure=?, pincode=?, employment_type=?, gross_monthly_income=? WHERE id=?",
// 		existingLead.Loan_type,
// 		existingLead.Loan_amount,
// 		existingLead.Tenure,
// 		existingLead.Pincode,
// 		existingLead.Employment_type,
// 		existingLead.Gross_monthly_income,
// 		// existingLead.Status,
// 		existingLead.ID,
// 	)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "Lead updated successfully")
// }

//////// test update

// func UpdateLeadTest(w http.ResponseWriter, r *http.Request) {
// 	config, err := con.LoadConfig("Config/config.yaml")
// 	if err != nil {
// 		panic(err)
// 	}

// 	con.ConnectDB(config)
// 	defer con.CloseDB()

// 	params := mux.Vars(r)
// 	LeadID := params["id"]

// 	// Check if the request method is allowed
// 	if r.Method != http.MethodPatch {
// 		w.Header().Set("Allow", http.MethodPatch)
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var lead Lead_info
// 	err = json.NewDecoder(r.Body).Decode(&lead)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	query := "UPDATE lead_table SET loan_type=?, loan_amount=?, tenure=?, pincode=?, employment_type=?, gross_monthly_income=? WHERE id=?"
// 	_, error := db.Query(query, lead.Loan_type, lead.Loan_amount, lead.Tenure, lead.Pincode, lead.Employment_type, lead.Gross_monthly_income, LeadID)
// 	if err != nil {
// 		http.Error(w, error.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Write the query to the HTTP response
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(query))

// 	// Print the query to the browser
// 	w.Write([]byte("\n"))
// 	w.Write([]byte("Query with Dynamic Data: " + formatQuery(query, lead.Loan_type, lead.Loan_amount, lead.Tenure, lead.Pincode, lead.Employment_type, lead.Gross_monthly_income, LeadID)))
// }

// func formatQuery(query string, data ...interface{}) string {
// 	for _, d := range data {
// 		query = strings.Replace(query, "?", fmt.Sprintf("'%v'", d), 1)
// 	}
// 	return query
// }
