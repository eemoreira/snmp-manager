package models

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
	ID         int    `db:"id"`
	MaquinaIP  string    `db:"maquina_ip"`
	ExecutarEm string `db:"executar_em"`
	Acao       string `db:"acao"`
	executado  bool   `db:"executado"`
}
