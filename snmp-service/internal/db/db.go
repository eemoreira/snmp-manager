package db

import (
	"github.com/eemoreira/snmp-manager/internal/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
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

func (m *DBManager) GetPortaByIP(ip string) (int, error) {
	var ps int
	query := `
	  SELECT ps.* FROM porta_switch ps
	  JOIN switch_ s ON s.id = ps.switch_id
	  WHERE s.ip = ?
	`
	err := m.DB.Get(&ps, query, ip)
	if err != nil {
		return -1, err
	}
	return ps, nil
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

func (m *DBManager) CreateAgendamento(maquina_ip string, ativo bool, execTime time.Time) error {
	_, err := m.DB.Exec(
		"INSERT INTO agendamento (maquina_ip, ativo, tempo) VALUES (?, ?, ?)",
		maquina_ip, ativo, execTime,
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
