package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"fmt"

	"github.com/eemoreira/snmp-manager/internal/db"
	"github.com/eemoreira/snmp-manager/internal/models"
	"github.com/eemoreira/snmp-manager/internal/snmp"
)

const (
	SWITCH_IP = "10.90.90.90"
	RW_COMM   = "private"
	RO_COMM   = "public"
)

type Handler struct {
	DB      *db.DBManager
	sala    *models.Sala
	Maquina *models.Maquina
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) setPortCtx(ip, state string) (error, int) {
	mgr := snmp.NewManager(SWITCH_IP, RW_COMM)
	err := mgr.Connect()
	if err != nil {
		return err, 500
	}
	defer mgr.Close()

	port, err := h.DB.GetPortaByIP(ip)
	if err != nil {
		return err, 500
	}

	if !strings.EqualFold(state, "up") && !strings.EqualFold(state, "down") {
		return fmt.Errorf("State must be 'up' or 'down'"), 400
	}

	up := strings.EqualFold(state, "up")

	ok, err := mgr.SetPortStatus(port, up)
	if err != nil || !ok {
		return err, 500
	}

	return nil, 200
}

func (h *Handler) setPort(w http.ResponseWriter, r *http.Request) {
	mgr := snmp.NewManager(SWITCH_IP, RW_COMM)
	err := mgr.Connect()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "Failed to connect to switch: " + err.Error()})
		return
	}
	defer mgr.Close()

	var req struct {
		IP    string `json:"ip"`
		State string `json:"state"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "Invalid request payload"})
		return
	}

	err, code := h.setPortCtx(req.IP, req.State)
	if err != nil {
		writeJSON(w, code, map[string]string{"error": "Failed to set port status: " + err.Error()})
		return
	}

	writeJSON(w, code, map[string]bool{"success": true})
}

func (h *Handler) getPort(w http.ResponseWriter, r *http.Request) {
	mgr := snmp.NewManager(SWITCH_IP, RO_COMM)
	err := mgr.Connect()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "Failed to connect to switch: " + err.Error()})
		return
	}
	defer mgr.Close()

	var req struct {
		IP string `json:"ip"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "Invalid request payload"})
		return
	}

	port, err := h.DB.GetPortaByIP(req.IP)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "Failed to get port for IP: " + err.Error()})
		return
	}

	status, err := mgr.GetPortStatus(port)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "Failed to get port status: " + err.Error()})
		return
	}

	writeJSON(w, 200, map[string]int{"status": status})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "Invalid request payload"})
		return
	}

	if !strings.EqualFold(creds.Login, h.sala.LoginAdmin) || !strings.EqualFold(creds.Password, h.sala.SenhaAdmin) {
		writeJSON(w, 401, map[string]string{"error": "Invalid credentials"})
		return
	}

	writeJSON(w, 200, map[string]string{"message": "Login successful"})
}


func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr

	isAdmin, sala, err := h.DB.IsIPAdmin(remoteIP)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "Database error: " + err.Error()})
		return
	}

	maquina, err := h.DB.GetMaquinaByIP(remoteIP)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "Database error: " + err.Error()})
		return
	}

	h.sala = sala
	h.Maquina = maquina

	if !isAdmin {
		writeJSON(w, 403, map[string]string{"error": "Access denied"})
		return
	}

	go h.agendamentoWorker()

	writeJSON(w, 200, map[string]string{"message": "Welcome to the SNMP Manager"})

}

func (h *Handler) createMaquina(w http.ResponseWriter, r *http.Request) {
	var mq struct {
		IP       string `json:"ip"`
		MAC      string `json:"mac"`
	}
	err := json.NewDecoder(r.Body).Decode(&mq)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "Invalid request payload"})
		return
	}
	
	id, err := h.DB.CreateMaquina(mq.IP, mq.MAC)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "Failed to create machine: " + err.Error()})
		return
	}

	writeJSON(w, 201, map[string]int64{"id": id})
}

func (h *Handler) createAgendamento(w http.ResponseWriter, r *http.Request) {

	var req struct {
		IP    string `json:"ip"`
		State string `json:"state"`
		TimeOffset  string `json:"time_offset"`
	}

	offset, err := time.ParseDuration(req.TimeOffset + "s")
	execTime := time.Now().Add(offset * time.Second)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "Invalid request payload"})
		return
	}

	if !strings.EqualFold(req.State, "up") && !strings.EqualFold(req.State, "down") {
		writeJSON(w, 400, map[string]string{"error": "State must be 'up' or 'down'"})
		return
	}

	up := strings.EqualFold(req.State, "up")
	err = h.DB.CreateAgendamento(req.IP, up, execTime)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "Failed to set schedule: " + err.Error()})
		return
	}

	writeJSON(w, 200, map[string]bool{"success": true})

}

func (h *Handler) agendamentoWorker() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		<-ticker.C
		agendamentos, err := h.DB.GetTODOAgendamentos()
		if err != nil {
			fmt.Println("Failed to fetch scheduled tasks:", err)
			continue
		}
		for _, ag := range agendamentos {
			mq, err := h.DB.GetMaquinaByIP(ag.MaquinaIP)
			if err != nil {
				fmt.Println("Failed to get machine for schedule:", err)
				continue
			}

			if !strings.EqualFold(ag.Acao, "up") && !strings.EqualFold(ag.Acao, "down") {
				fmt.Println("Invalid action in schedule:", ag.Acao)
				continue
			}
			
			err, _ = h.setPortCtx(mq.IP, ag.Acao)
			if err != nil {
				fmt.Println("Failed to execute scheduled action:", err)
				continue
			}

			err = h.DB.MarkAgendamentoExecuted(ag.ID)
			if err != nil {
				fmt.Println("Failed to mark schedule as executed:", err)
				continue
			}
		}
	}
}

