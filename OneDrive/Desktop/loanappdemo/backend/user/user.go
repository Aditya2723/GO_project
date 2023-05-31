package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/go-playground/validator"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type User_info struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Contact_Num   int       `json:"contact"`
	Password      string    `json:"password,omitempty"`
	Credit_score  int       `json:"credit_score"`
	Created_At    time.Time `json:"created at"`
	Last_Modified time.Time `json:"last modified"`
}

type Error struct {
	Message string `json:"message"`
}
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error *Error      `json:"error,omitempty"`
}

func isEmailValid(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	return emailRegex.MatchString(email)
}

func ValidateUser(user User_info, validate *validator.Validate) error {
	// Check that user name is not empty
	if user.Name == "" {
		return errors.New("user name is required.")
	}
	// Check that phone no. is exactly 10 digits
	if match, _ := regexp.MatchString(`^\d{10}$`, fmt.Sprint(user.Contact_Num)); !match {
		return errors.New("Phone number must be exactly 10 digits.")
	}

	// Check that password is not empty
	if user.Password == "" {
		return errors.New("Password is required.")
	}
	// Check that email is valid
	if !isEmailValid(user.Email) {
		return errors.New("Invalid email format.")
	}
	// If all checks pass, return nill
	return nil
}

func dbConnect() (db *sql.DB) {
	db, err := sql.Open("mysql", "root:7572@tcp(127.0.0.1:3306)/loan_application")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("database connected")
	return db
}
func GetUserById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]
	db := dbConnect()
	var user User_info
	err := db.QueryRow("SELECT id, user_name, user_email, user_contact_num, credit_score FROM user_info WHERE id=?", userID).Scan(&user.Id, &user.Name, &user.Email, &user.Contact_Num, &user.Credit_score)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(userJSON)
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	var user User_info
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()

	// validate loan details
	if err := ValidateUser(user, validate); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&Response{Error: &Error{Message: err.Error()}})
		return
	}
	db := dbConnect()
	result, err := db.Exec("INSERT INTO user_info (user_name, user_email, user_contact_num, user_password,credit_score,created_at,last_modified) VALUE(?, ?, ?, ?,?, NOW(), NOW())", user.Name, user.Email, user.Contact_Num, user.Password, user.Credit_score)
	if err != nil {
		log.Fatal(err)
	}
	id, _ := result.LastInsertId()
	fmt.Fprintf(w, "New student has been created with ID: %d", id)
}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]
	db := dbConnect()
	result, err := db.Exec("DELETE FROM user_info WHERE id=?", userID)
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
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User deleted successfully")
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]
	db := dbConnect()
	var user User_info
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := db.Exec("UPDATE user_info SET user_name=?, user_email=?, user_contact_num=?, user_password=?, last_modified=NOW() WHERE id=?", user.Name, user.Email, user.Contact_Num, user.Password, userID)
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
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User updated successfully")
}
