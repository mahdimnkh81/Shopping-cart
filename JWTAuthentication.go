package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

const secretKey = "mahdi"

func Signup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string
		Password string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error decoding request body into CreateBasketBody struct %v", err)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Failed to hash password %v", err)
		return
	}

	_, err = DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", body.Username, string(hash))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Failed to Create user, Try again %v", err)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) error {

	var body struct {
		Username string
		Password string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		//log.Printf("error decoding request body into CreateBasketBody struct %v", err)
		return errors.New("error decoding request body into CreateBasketBody struct")
	}

	var user User
	if err := DB.QueryRow("SELECT username, password FROM public.users WHERE username=$1", body.Username).
		Scan(&user.Username, &user.Password); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("user not found")
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("invalid username or password")
	}

	//	Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Username,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secretKey))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("invalid create token")
	}

	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   3600 * 24 * 30,
		HttpOnly: true,
		Secure:   false,
	}
	http.SetCookie(w, &cookie)
	w.Write([]byte("cookie set!"))
	log.Println(err)
	return nil
}

func Authorize(w http.ResponseWriter, r *http.Request) (error, string) {
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			http.Error(w, "cookie not found", http.StatusBadRequest)
		default:
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return errors.New("cookie not found, please login"), ""
	}

	tokenString := cookie.Value

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		log.Fatal(err)
	}

	var username string
	if claims, ok := token.Claims.(jwt.MapClaims); ok {

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return errors.New("token has expired, please login"), ""
		}
		var user User
		if err := DB.QueryRow("SELECT username, password FROM public.users WHERE username=$1", claims["sub"]).
			Scan(&user.Username, &user.Password); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return errors.New("user not found"), ""
		}
		if user.Username == "" {
			return errors.New("invalid Token, please login"), ""
		}
		username = claims["sub"].(string)
	} else {
		return errors.New("user not found"), ""
	}

	return nil, username
}
