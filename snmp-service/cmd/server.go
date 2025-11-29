package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/eemoreira/snmp-manager/internal/db"
	"github.com/gorilla/mux"
)

func main() {

	dsn := os.Getenv("DB_DSN")
	database, err := db.NewDBManager(dsn)
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		os.Exit(1)
	}
	defer database.Close()

	h := &Handler{DB: database}
	go h.agendamentoWorker()

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	r.HandleFunc("/api/check", h.check).Methods("GET")
	r.HandleFunc("/api/login", h.login).Methods("POST")
	r.HandleFunc("/api/maquinas", h.createMaquina).Methods("POST")

	r.HandleFunc("/api/agendamento", h.createAgendamento).Methods("POST")
	r.HandleFunc("/api/ports", h.setPort).Methods("POST")
	r.HandleFunc("/api/ports", h.getPort).Methods("GET")

	http.ListenAndServe(":8080", r)
}
