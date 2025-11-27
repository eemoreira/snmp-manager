package main 

import (
    "net/http"
	"fmt"
	"os"

	"github.com/eemoreira/snmp-manager/internal/db"
    "github.com/gorilla/mux"
)


func main() {

	dsn := os.Getenv("DSN")
	database, err := db.NewDBManager(dsn)
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		os.Exit(1)
	}
	defer database.Close()

	h := &Handler{DB: database}

    r := mux.NewRouter()
	r.HandleFunc("/", h.home).Methods("GET")
	r.HandleFunc("/login", h.login).Methods("POST")
	r.HandleFunc("/api/maquinas", h.createMaquina).Methods("POST")
	r.HandleFunc("/api/agendamento", h.createAgendamento).Methods("POST")
    r.HandleFunc("/api/ports", h.setPort).Methods("POST")
    r.HandleFunc("/api/ports", h.getPort).Methods("GET")

    http.ListenAndServe(":8080", r)
}

