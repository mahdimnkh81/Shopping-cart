package main

import "log"

func main() {
	err := OpenDataBase()
	if err != nil {
		log.Printf("error connectiong to postgress database %v", err)
	}
	defer CloseDatabase()

	server := NewAPIServer(":3000")
	server.Run()
}
