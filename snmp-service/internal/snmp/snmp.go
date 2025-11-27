package snmp 

import (
	"fmt"
	"github.com/gosnmp/gosnmp"
	"time"
)

type Manager struct {
	Client *gosnmp.GoSNMP
}

func NewManager(target string, community string) *Manager {
	client := &gosnmp.GoSNMP{
		Target:    target,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version1,
		Timeout:   time.Duration(2) * time.Second,
	}
	return &Manager{Client: client}
}

func (m *Manager) Connect() error {
	return m.Client.Connect()
}

func (m *Manager) Close() {
	m.Client.Conn.Close()
}

func (m *Manager) SetPortStatus(port int, up bool) (bool, error) {
	oid := fmt.Sprintf("1.3.6.1.2.1.2.2.1.7.%d", port)
	status := 2
	if up {
		status = 1
	}
	pdu := gosnmp.SnmpPDU{
		Name:  oid,
		Type:  gosnmp.Integer,
		Value: status,
	}
	_, err := m.Client.Set([]gosnmp.SnmpPDU{pdu})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *Manager) GetPortStatus(port int) (int, error) {
	oid := fmt.Sprintf("1.3.6.1.2.1.2.2.1.8.%d", port)
	result, err := m.Client.Get([]string{oid})
	if err != nil {
		return 0, err
	}
	if len(result.Variables) == 0 {
		return 0, fmt.Errorf("no variables returned")
	}
	v := result.Variables[0]
	// Expect integer
	switch intval := v.Value.(type) {
	case int:
		return intval, nil
	case uint:
		return int(intval), nil
	case int64:
		return int(intval), nil
	default:
		return 0, fmt.Errorf("unexpected type %T", v.Value)
	}
}
