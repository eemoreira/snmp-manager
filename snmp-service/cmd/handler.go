package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/eemoreira/snmp-manager/internal/db"
	"github.com/eemoreira/snmp-manager/internal/snmp"
	"github.com/gorilla/mux"
)

const (
	SWITCH_IP = "10.90.90.90"
	RW_COMM  = "private"
	RO_COMM  = "public"
)

type Handler struct {
	DB *db.DBManager
}

func (h *Handler) setPort(w http.ResponseWriter, r *http.Request) {
	mgr := snmp.NewManager(SWITCH_IP, RW_COMM)
	err := mgr.Connect()
	if err != nil {
		http.Error(w, "Failed to connect to switch: "+err.Error(), 500)
		return
	}
	defer mgr.Close()
    params := mux.Vars(r)
    port, err := strconv.Atoi(params["port"])
    if err != nil {
        http.Error(w, "Invalid port", 400)
        return
    }

	upParam := params["state"]
	if upParam != "up" && upParam != "down" {
		http.Error(w, "invalid state", 400);
	}
    up := upParam == "up"

    ok, err := mgr.SetPortStatus(port, up)
    if err != nil || !ok {
        http.Error(w, "Failed to set port: "+err.Error(), 500)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) getPort(w http.ResponseWriter, r *http.Request) {
	mgr := snmp.NewManager(SWITCH_IP, RO_COMM)
	err := mgr.Connect()
	if err != nil {
		http.Error(w, "Failed to connect to switch: "+err.Error(), 500)
		return
	}
	defer mgr.Close()
    params := mux.Vars(r)
    port, err := strconv.Atoi(params["port"])
    if err != nil {
        http.Error(w, "Invalid port", 400)
        return
    }

    status, err := mgr.GetPortStatus(port)
    if err != nil {
        http.Error(w, "Failed to get port: "+err.Error(), 500)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]int{"status": status})
}
