package main

import (
	"encoding/json"
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

func AA() {
	fmt.Println("mahdi")
}

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/basket", makeHTTPHandleFunc(s.handleGetAllBasket))
	router.HandleFunc("/basket/{id}", makeHTTPHandleFunc(s.handleBasket))
	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleBasket(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetBasketById(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreatBasket(w, r)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteBasketById(w, r)
	}
	if r.Method == "PATCH" {
		return s.handleUpdateBasket(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleGetAllBasket(w http.ResponseWriter, r *http.Request) error {
	var getbaskets []GetBasket

	rows, err := DB.Query("SELECT id, username, created_at, updated_at, data, state FROM public.baskets;")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	for rows.Next() {
		var getbasket GetBasket
		if err := rows.Scan(&getbasket.ID, &getbasket.Username, &getbasket.CreatedAt, &getbasket.UpdatedAt, &getbasket.Data, &getbasket.State); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		getbaskets = append(getbaskets, getbasket)
	}
	_, err = json.Marshal(getbaskets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error marshaling")
		return err
	}
	//w.Write(j)
	return WriteJSON(w, http.StatusOK, getbaskets)

}
func (s *APIServer) handleCreatBasket(w http.ResponseWriter, r *http.Request) error {

	var body CreateBasket
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error decoding request body into CreateBasketBody struct %v", err)
		return err
	}
	createdAt := time.Now().Format(time.RFC3339)
	updatedAt := createdAt
	if err := DB.QueryRow("INSERT INTO baskets (username,created_at, updated_at, data, state) VALUES ($1, $2, $3, $4,$5)",
		body.Username, createdAt, updatedAt, body.Data, body.State).Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return WriteJSON(w, http.StatusOK, "Succes")

}

func (s *APIServer) handleGetBasketById(w http.ResponseWriter, r *http.Request) error {

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
	//w.Write(j)
	return WriteJSON(w, http.StatusOK, getbasket)

}
func (s *APIServer) handleDeleteBasketById(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}
	err = DB.QueryRow("delete from public.baskets where id = $1", id).Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	//_, err = json.Marshal(getbasket)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	log.Printf("error marshaling")
	//	return err
	//}
	//w.Write(j)
	return WriteJSON(w, http.StatusOK, "Delete successfully")

}
func (s *APIServer) handleUpdateBasket(w http.ResponseWriter, r *http.Request) error {

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}

	var body CreateBasket
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error decoding request body into CreateBasketBody struct %v", err)
		return err
	}

	var getbasket GetBasket
	if err := DB.QueryRow("SELECT id, username, created_at, updated_at, data, state FROM public.baskets WHERE id=$1", id).
		Scan(&getbasket.ID, &getbasket.Username, &getbasket.CreatedAt, &getbasket.UpdatedAt, &getbasket.Data, &getbasket.State); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	if getbasket.State == 1 {
		return fmt.Errorf("can Not Update")
	} else {
		getbasket.UpdatedAt = time.Now().Format(time.RFC3339)
		getbasket.Data = body.Data
	}

	return WriteJSON(w, http.StatusOK, getbasket)

}
