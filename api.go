package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

type APIServer struct {
	listenAddr string
}

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() {

	router := mux.NewRouter()
	router.HandleFunc("/Signup", Signup).Methods("POST")
	router.HandleFunc("/Login", makeHTTPHandleFunc(Login)).Methods("POST")
	router.HandleFunc("/Authorize", s.handleAuthorize).Methods("POST")
	router.HandleFunc("/basket", makeHTTPHandleFunc(s.handleCreatBasket)).Methods("POST")
	router.HandleFunc("/basket/{id}", makeHTTPHandleFunc(s.handleGetBasketById)).Methods("GET")
	router.HandleFunc("/basket", makeHTTPHandleFunc(s.handleGetAllBasket)).Methods("GET")
	router.HandleFunc("/basket/{id}", makeHTTPHandleFunc(s.handleDeleteBasketById)).Methods("DELETE")
	router.HandleFunc("/basket/{id}", makeHTTPHandleFunc(s.handleUpdateBasket)).Methods("PATCH")

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	Authorize(w, r)
}

func (s *APIServer) handleGetAllBasket(w http.ResponseWriter, r *http.Request) error {

	errAuth, username := Authorize(w, r)
	if errAuth != nil {
		return errAuth
	}

	var getbaskets []GetBasket

	rows, err := DB.Query("SELECT id, username, created_at, updated_at, data, state FROM public.baskets;")
	if err != nil && err.Error() != "EOF" {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("you can't see this")
	}
	for rows.Next() {
		var getbasket GetBasket
		if err := rows.Scan(&getbasket.ID, &getbasket.Username, &getbasket.CreatedAt, &getbasket.UpdatedAt, &getbasket.Data, &getbasket.State); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		if username == getbasket.Username {
			getbaskets = append(getbaskets, getbasket)
		}
	}
	_, err = json.Marshal(getbaskets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error marshaling")
		return err
	}
	if len(getbaskets) == 0 {
		return WriteJSON(w, http.StatusOK, "you don't have any basket to see")
	} else {
		return WriteJSON(w, http.StatusOK, getbaskets)
	}
}
func (s *APIServer) handleCreatBasket(w http.ResponseWriter, r *http.Request) error {

	errAuth, username := Authorize(w, r)
	if errAuth != nil {
		return errAuth
	}

	var body CreateBasket
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err.Error() != "EOF" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error decoding request body into CreateBasketBody struct %v", err)
		return err
	}
	createdAt := time.Now().Format(time.RFC3339)
	updatedAt := createdAt
	/*fmt.Println(body.Data)*/

	if err := DB.QueryRow("INSERT INTO baskets (username,created_at, updated_at, data, state) VALUES ($1, $2, $3, $4,$5)",
		//username come from token
		username, createdAt, updatedAt, body.Data, body.State).Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return WriteJSON(w, http.StatusOK, "Succes")

}

func (s *APIServer) handleGetBasketById(w http.ResponseWriter, r *http.Request) error {

	errAuth, username := Authorize(w, r)
	if errAuth != nil {
		return errAuth
	}
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}
	var getbasket GetBasket
	if err := DB.QueryRow("SELECT id, username, created_at, updated_at, data, state FROM public.baskets WHERE id=$1", id).
		Scan(&getbasket.ID, &getbasket.Username, &getbasket.CreatedAt, &getbasket.UpdatedAt, &getbasket.Data, &getbasket.State); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	_, err = json.Marshal(getbasket)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error marshaling")
		return err
	}
	if username == getbasket.Username {
		return WriteJSON(w, http.StatusOK, getbasket)
	} else {
		return WriteJSON(w, http.StatusOK, "you don't have any basket to see or don't have access")
	}

}
func (s *APIServer) handleDeleteBasketById(w http.ResponseWriter, r *http.Request) error {
	errAuth, username := Authorize(w, r)
	if errAuth != nil {
		return errAuth
	}
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}
	var getbasket GetBasket
	if err := DB.QueryRow("SELECT id, username, created_at, updated_at, data, state FROM public.baskets WHERE id=$1", id).
		Scan(&getbasket.ID, &getbasket.Username, &getbasket.CreatedAt, &getbasket.UpdatedAt, &getbasket.Data, &getbasket.State); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	if username == getbasket.Username {
		err = DB.QueryRow("delete from public.baskets where id = $1", id).Err()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		return WriteJSON(w, http.StatusOK, "Delete successfully")
	} else {
		return WriteJSON(w, http.StatusOK, "you don't have access to delete")
	}

}
func (s *APIServer) handleUpdateBasket(w http.ResponseWriter, r *http.Request) error {

	errAuth, username := Authorize(w, r)
	if errAuth != nil {
		return errAuth
	}
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}

	var body CreateBasket
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err.Error() != "EOF" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error decoding request body into UpdateBasketBody struct %v", err)
		return err
	}

	var getbasket GetBasket
	if err := DB.QueryRow("SELECT id, username, created_at, updated_at, data, state FROM public.baskets WHERE id=$1", id).
		Scan(&getbasket.ID, &getbasket.Username, &getbasket.CreatedAt, &getbasket.UpdatedAt, &getbasket.Data, &getbasket.State); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	if username != getbasket.Username {
		return WriteJSON(w, http.StatusOK, "you don't have any basket to see or don't have access")
	}
	if getbasket.State == 1 {
		return fmt.Errorf("can Not Update because state is COMPLETED")
	} else {
		getbasket.UpdatedAt = time.Now().Format(time.RFC3339)
		getbasket.Data = body.Data
		getbasket.State = body.State

		_, err := DB.Exec("UPDATE baskets SET data = $1, state = $2, updated_at = $3 WHERE id = $4", getbasket.Data, getbasket.State, getbasket.UpdatedAt, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		return WriteJSON(w, http.StatusOK, getbasket)
	}
}
