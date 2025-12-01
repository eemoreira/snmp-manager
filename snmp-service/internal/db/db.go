package db

import (
	"time"

	"github.com/eemoreira/snmp-manager/internal/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type DBManager struct {
	DB *sqlx.DB
}

func NewDBManager(dsn string) (*DBManager, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &DBManager{DB: db}, nil
}

func (m *DBManager) Close() {
	m.DB.Close()
}

func (m *DBManager) CreateMaquina(ip, mac string) (int64, error) {
	res, err := m.DB.Exec(
		"INSERT INTO maquina (ip, mac) VALUES (?, ?)",
		ip, mac,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (m *DBManager) LinkMaquinaToPortaSwitch(maquinaIP string, portaSwitchID int) error {
	_, err := m.DB.Exec(
		"INSERT INTO sala_porta (ip_maquina, porta_switch_id) VALUES (?, ?)",
		maquinaIP, portaSwitchID,
	)
	return err
}

func (m *DBManager) GetPortaNumByMaquinaIP(maquinaIP string) (int, error) {
	var portaNumero int
	query := `
        SELECT ps.porta_numero 
        FROM porta_switch ps
        JOIN sala_porta sp ON sp.porta_switch_id = ps.id
        WHERE sp.ip_maquina = ?
    `
	err := m.DB.Get(&portaNumero, query, maquinaIP)
	if err != nil {
		return -1, err
	}
	return portaNumero, nil
}

func (m *DBManager) GetMaquinaByIP(ip string) (*models.Maquina, error) {
	var mq models.Maquina
	err := m.DB.Get(&mq, "SELECT * FROM maquina WHERE ip = ?", ip)
	if err != nil {
		return nil, err
	}
	return &mq, nil
}

func (m *DBManager) CreateSwitch(ip string) (int64, error) {
	res, err := m.DB.Exec(
		"INSERT INTO switch_ (ip) VALUES (?)",
		ip,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type PortaSwitch struct {
	ID          int `db:"id"`
	SwitchID    int `db:"switch_id"`
	PortaNumero int `db:"porta_numero"`
}

func (m *DBManager) CreatePortaSwitch(switchID, portaNum int) (int64, error) {
	res, err := m.DB.Exec(
		"INSERT INTO porta_switch (switch_id, porta_numero) VALUES (?, ?)",
		switchID, portaNum,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *DBManager) CreateSala(nome string, adminMaquinaID int, login, senha string) (int64, error) {
	res, err := m.DB.Exec(
		"INSERT INTO sala (nome, maquina_admin_id, login_admin, senha_admin) VALUES (?, ?, ?, ?)",
		nome, adminMaquinaID, login, senha,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *DBManager) CreateSalaPorta(salaID, portaSwitchID, ip string) (int64, error) {
	res, err := m.DB.Exec(
		"INSERT INTO sala_porta (sala_id, porta_switch_id, ip_maquina) VALUES (?, ?, ?)",
		salaID, portaSwitchID, ip,
	)
	if err != nil {
		return 0, nil
	}
	return res.LastInsertId()
}

func (m *DBManager) GetSalaFromPortaSwitch(switchIP string, portaNum int) (*models.Sala, error) {
	var sala models.Sala
	query := `
        SELECT s.* FROM sala s
        JOIN sala_porta sp ON sp.sala_id = s.id
        JOIN porta_switch ps ON ps.id = sp.porta_switch_id
        JOIN switch_ sw ON sw.id = ps.switch_id
        WHERE sw.ip = ? AND ps.porta_numero = ?
    `
	err := m.DB.Get(&sala, query, switchIP, portaNum)
	if err != nil {
		return nil, err
	}
	return &sala, nil
}

func (m *DBManager) GetPortaSwitchID(switchIP string, portaNum int) (int, error) {
	var portaSwitchID int
	query := `
		SELECT ps.id FROM porta_switch ps
		JOIN switch_ sw ON sw.id = ps.switch_id
		WHERE sw.ip = ? AND ps.porta_numero = ?
	`
	err := m.DB.Get(&portaSwitchID, query, switchIP, portaNum)
	if err != nil {
		return -1, err
	}
	return portaSwitchID, nil
}

func (m *DBManager) IsIPAdmin(ip string) (bool, *models.Sala, error) {
	var sala models.Sala
	query := `
      SELECT s.* FROM sala s
      JOIN maquina m ON m.id = s.maquina_admin_id
      WHERE m.ip = ?
    `
	err := m.DB.Get(&sala, query, ip)
	if err != nil {
		return false, nil, err
	}
	return true, &sala, nil
}

// É necessário passar o salaID e converter o boolean para o ENUM correto
func (m *DBManager) CreateAgendamento(salaID int, maquinaIP string, acao string, execTime time.Time) error {
	// Validação simples para garantir que a string combine com o ENUM do banco

	_, err := m.DB.Exec(
		"INSERT INTO agendamento (sala_id, ip_maquina, acao, executar_em, executado) VALUES (?, ?, ?, ?, ?)",
		salaID, maquinaIP, acao, execTime, false,
	)
	return err
}

func (m *DBManager) GetTODOAgendamentos() ([]models.Agendamento, error) {
	var agendamentos []models.Agendamento
	err := m.DB.Select(&agendamentos, "SELECT * FROM agendamento WHERE executado = 0 and executar_em <= ?", time.Now())
	if err != nil {
		return nil, err
	}
	return agendamentos, nil
}

func (m *DBManager) MarkAgendamentoExecuted(id int) error {
	_, err := m.DB.Exec(
		"UPDATE agendamento SET executado = 1 WHERE id = ?", id,
	)
	return err
}

func (m *DBManager) LinkPortaToSala(portaSwitchID int, salaID int) error {
	// Verificar se já existe a associação
	var count int
	err := m.DB.Get(&count, "SELECT COUNT(*) FROM sala_porta WHERE porta_switch_id = ? AND sala_id = ?", portaSwitchID, salaID)
	if err != nil {
		return err
	}

	// Se já existe, não fazer nada
	if count > 0 {
		return nil
	}

	// Inserir a associação
	_, err = m.DB.Exec(
		"INSERT INTO sala_porta (sala_id, porta_switch_id) VALUES (?, ?)",
		salaID, portaSwitchID,
	)
	return err
}
