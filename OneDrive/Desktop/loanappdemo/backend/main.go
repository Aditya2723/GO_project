package main

import (
	con "backend/Config"
	"backend/lead"
	"backend/loan"
	"html/template"

	"backend/user"

	"fmt"

	"log"
	"net/http"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
)

var tmpl = template.Must(template.ParseGlob("form/*.html"))

func TemplatePage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "testlead.html", nil)
}

func main() {

	// Load the config file

	config, err := con.LoadConfig("./Config/config.yaml")
	if err != nil {
		panic(err)
	}

	con.ConnectDB(config)
	defer con.CloseDB()

	r := mux.NewRouter()

	r.PathPrefix("/form/").Handler(http.StripPrefix("/form/", http.FileServer(http.Dir("form")))) //localhost:9000/form/ endpoint for accessing forms
	r.HandleFunc("/", TemplatePage).Methods("GET")
	r.HandleFunc("/lead/get/{id}", lead.LeadIndexAll).Methods("GET")
	r.HandleFunc("/lead/add", lead.InsertLead).Methods("POST")
	r.HandleFunc("/lead/update/{id}", lead.UpdateLead).Methods("PATCH")
	r.HandleFunc("/lead/delete/{id}", lead.DeleteLead).Methods("DELETE")

	//Admin Panel's endpoint
	r.HandleFunc("/lead/admin", lead.LeadIndex).Methods("GET")

	//r.HandleFunc("/lead/upd/{id}", lead.UpdateLeadTest).Methods("PATCH")
	//r.HandleFunc("/lead/admin/update/{id}", lead.UpdateLeadDashboard).Methods("PATCH")
	//r.HandleFunc("/lead/handle/{id}", lead.HandleUpdateLead).Methods("PATCH")

	//////////////////////////////////////////////////////////////////////
	r.HandleFunc("/loan/index/{id}", loan.GetLoanDetails).Methods("GET")
	r.HandleFunc("/loan/insert", loan.InsertLoanDetails).Methods("POST")
	r.HandleFunc("/loan/update/{id}", loan.UpdateLoanDetails).Methods("PATCH")
	r.HandleFunc("/loan/delete/{id}", loan.DeleteLoanDetails).Methods("DELETE")

	//all loan details for admin panel
	r.HandleFunc("/loan/get", loan.LoanIndex).Methods("GET")

	r.HandleFunc("/user/add", user.AddUser).Methods("POST")
	r.HandleFunc("/user/get/{id}", user.GetUserById).Methods("GET")
	r.HandleFunc("/user/update/{id}", user.UpdateUser).Methods("PATCH")
	r.HandleFunc("/user/delete/{id}", user.DeleteUser).Methods("DELETE")

	r.HandleFunc("/user/refresh", user.Refresh).Methods("GET")
	r.HandleFunc("/user/home", user.Home).Methods("GET")
	r.HandleFunc("/user/login", user.Login).Methods("POST")

	fmt.Println("Server started at http://localhost:9000")
	log.Fatal(http.ListenAndServe(":9000", r))

}
