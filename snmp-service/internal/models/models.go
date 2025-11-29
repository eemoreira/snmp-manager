package models
import "time"

type Sala struct {
	ID             int    `db:"id"`
	Nome           string `db:"nome"`
	MaquinaAdminID int    `db:"maquina_admin_id"`
	LoginAdmin     string `db:"login_admin"`
	SenhaAdmin     string `db:"senha_admin"`
}

type Maquina struct {
	ID  int    `db:"id"`
	IP  string `db:"ip"`
	MAC string `db:"mac"`
}

type Switch struct {
	ID int    `db:"id"`
	IP string `db:"ip"`
}

type Agendamento struct {
    ID          int       `db:"id"`
    SalaID      int       `db:"sala_id"`
    MaquinaIP   string    `db:"ip_maquina"`
    Acao        string    `db:"acao"`
    ExecutarEm  time.Time `db:"executar_em"`
    Executado   bool      `db:"executado"`
}
