package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte("secret key")

type Credential struct {
	Username string `json:"name,omitempty"`
	Password string `json:"password"`
	jwt.StandardClaims
}

type Session struct {
	ID        int       `json:"id,omitempty"`
	UserID    int       `json:"user_id,omitempty"`
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST")

	// Check request method
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var credential Credential
	err := json.NewDecoder(r.Body).Decode(&credential)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := GetUserFromDatabase(credential.Username)
	fmt.Println(user)
	if err != nil || user.Password != credential.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(30 * time.Minute)
	claims := &Credential{
		Username: credential.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session := &Session{
		UserID:    user.Id,
		Token:     tokenString,
		ExpiresAt: expirationTime,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = SaveSession(session)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	w.WriteHeader(http.StatusOK)
}

func GetUserFromDatabase(username string) (*User_info, error) {
	db, err := sql.Open("mysql", "root:7572@tcp(127.0.0.1:3306)/loan_application")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}
	defer db.Close()

	query := "SELECT id,user_name, user_password FROM user_info WHERE user_name = ? LIMIT 1"
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare the query: %v", err)
	}
	defer stmt.Close()

	user := &User_info{}
	err = stmt.QueryRow(username).Scan(&user.Name, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user from the database: %v", err)
	}

	return user, nil
}
func Home(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenString := cookie.Value
	token, err := jwt.ParseWithClaims(tokenString, &Credential{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*Credential)
	if !ok || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	username := claims.Username

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome, %s!", username)
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	// w.Write([]byte(cookie.String()))
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tokenString := cookie.Value
	token, err := jwt.ParseWithClaims(tokenString, &Credential{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*Credential)
	if !ok || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 5*time.Minute {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expirationTime := time.Now().Add(30 * time.Minute)
	newClaims := &Credential{
		Username: claims.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := newToken.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   newTokenString,
		Expires: expirationTime,
	})

	w.WriteHeader(http.StatusOK)

}

func SaveSession(session *Session) error {
	db, err := sql.Open("mysql", "root:7572@tcp(127.0.0.1:3306)/loan_application")
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %v", err)
	}
	defer db.Close()
	user := &User_info{}
	// Fetch the user ID from the user_info table
	userID, err := GetUserIDFromDatabase(user.Name)
	if err != nil {
		return fmt.Errorf("failed to fetch user ID from the database: %v", err)
	}

	query := "INSERT INTO sessions (UserID, token, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"

	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare the query: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, session.Token, session.ExpiresAt, session.CreatedAt, session.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save session in the database: %v", err)
	}

	return nil
}

func GetUserIDFromDatabase(username string) (int, error) {
	db, err := sql.Open("mysql", "root:7572@tcp(127.0.0.1:3306)/loan_application")
	if err != nil {
		return 0, fmt.Errorf("failed to connect to the database: %v", err)
	}
	defer db.Close()

	query := "SELECT id FROM user_info WHERE user_name = ? LIMIT 1"
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare the query: %v", err)
	}
	defer stmt.Close()

	var userID int
	err = stmt.QueryRow(username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("failed to fetch user ID from the database: %v", err)
	}

	return userID, nil
}
