package db

import (
    "github.com/jmoiron/sqlx"
    _ "github.com/go-sql-driver/mysql"
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

type Maquina struct {
    ID       int    `db:"id"`
    IP       string `db:"ip"`
    MAC      string `db:"mac"`
    Hostname string `db:"hostname"`
}

func (m *DBManager) CreateMaquina(ip, mac, hostname string) (int64, error) {
    res, err := m.DB.Exec(
        "INSERT INTO maquina (ip, mac, hostname) VALUES (?, ?, ?)",
        ip, mac, hostname,
    )
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

func (m *DBManager) GetMaquinaByIP(ip string) (*Maquina, error) {
    var mq Maquina
    err := m.DB.Get(&mq, "SELECT * FROM maquina WHERE ip = ?", ip)
    if err != nil {
        return nil, err
    }
    return &mq, nil
}

type Switch struct {
    ID           int    `db:"id"`
    IP           string `db:"ip"`
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
    ID         int `db:"id"`
    SwitchID   int `db:"switch_id"`
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

type Sala struct {
    ID            int    `db:"id"`
    Nome          string `db:"nome"`
    MaquinaAdminID int   `db:"maquina_admin_id"`
    LoginAdmin    string `db:"login_admin"`
    SenhaAdmin    string `db:"senha_admin"`
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

func (m *DBManager) IsIPAdmin(ip string) (bool, *Sala, error) {
    var sala Sala
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

