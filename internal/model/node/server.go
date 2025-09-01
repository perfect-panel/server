package node

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Server struct {
	Id             int64      `gorm:"primary_key"`
	Name           string     `gorm:"type:varchar(100);not null;default:'';comment:Server Name"`
	Country        string     `gorm:"type:varchar(128);not null;default:'';comment:Country"`
	City           string     `gorm:"type:varchar(128);not null;default:'';comment:City"`
	Ratio          float32    `gorm:"type:DECIMAL(4,2);not null;default:0;comment:Traffic Ratio"`
	Address        string     `gorm:"type:varchar(100);not null;default:'';comment:Server Address"`
	Sort           int        `gorm:"type:int;not null;default:0;comment:Sort"`
	Protocols      string     `gorm:"type:text;default:null;comment:Protocol"`
	LastReportedAt *time.Time `gorm:"comment:Last Reported Time"`
	CreatedAt      time.Time  `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt      time.Time  `gorm:"comment:Update Time"`
}

func (*Server) TableName() string {
	return "servers"
}

func (m *Server) BeforeCreate(tx *gorm.DB) error {
	if m.Sort == 0 {
		var maxSort int
		if err := tx.Model(&Server{}).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort).Error; err != nil {
			return err
		}
		m.Sort = maxSort + 1
	}
	return nil
}

// MarshalProtocols Marshal server protocols to json
func (m *Server) MarshalProtocols(list []Protocol) error {
	var validate = make(map[string]bool)
	for _, protocol := range list {
		if protocol.Type == "" {
			return errors.New("protocol type is required")
		}
		if _, exists := validate[protocol.Type]; exists {
			return errors.New("duplicate protocol type: " + protocol.Type)
		}
		validate[protocol.Type] = true
	}
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	m.Protocols = string(data)
	return nil
}

// UnmarshalProtocols Unmarshal server protocols from json
func (m *Server) UnmarshalProtocols() ([]Protocol, error) {
	var list []Protocol
	if m.Protocols == "" {
		return list, nil
	}
	err := json.Unmarshal([]byte(m.Protocols), &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

type Protocol struct {
	Type                 string `json:"type"`
	Port                 uint16 `json:"port"`
	Security             string `json:"security,omitempty"`
	SNI                  string `json:"sni,omitempty"`
	AllowInsecure        bool   `json:"allow_insecure,omitempty"`
	Fingerprint          string `json:"fingerprint,omitempty"`
	RealityServerAddr    string `json:"reality_server_addr,omitempty"`
	RealityServerPort    int    `json:"reality_server_port,omitempty"`
	RealityPrivateKey    string `json:"reality_private_key,omitempty"`
	RealityPublicKey     string `json:"reality_public_key,omitempty"`
	RealityShortId       string `json:"reality_short_id,omitempty"`
	Transport            string `json:"transport,omitempty"`
	Host                 string `json:"host,omitempty"`
	Path                 string `json:"path,omitempty"`
	ServiceName          string `json:"service_name,omitempty"`
	Cipher               string `json:"cipher,omitempty"`
	ServerKey            string `json:"server_key,omitempty"`
	Flow                 string `json:"flow,omitempty"`
	HopPorts             string `json:"hop_ports,omitempty"`
	HopInterval          int    `json:"hop_interval,omitempty"`
	ObfsPassword         string `json:"obfs_password,omitempty"`
	DisableSNI           bool   `json:"disable_sni,omitempty"`
	ReduceRtt            bool   `json:"reduce_rtt,omitempty"`
	UDPRelayMode         string `json:"udp_relay_mode,omitempty"`
	CongestionController string `json:"congestion_controller,omitempty"`
}

// Marshal protocol to json
func (m *Protocol) Marshal() ([]byte, error) {
	type Alias Protocol
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	})
}

// Unmarshal json to protocol
func (m *Protocol) Unmarshal(data []byte) error {
	type Alias Protocol
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	return json.Unmarshal(data, &aux)
}
