package server

import (
	"time"

	"github.com/perfect-panel/server/pkg/logger"

	"gorm.io/gorm"
)

const (
	RelayModeNone        = "none"
	RelayModeAll         = "all"
	RelayModeRandom      = "random"
	RuleGroupTypeReject  = "reject"
	RuleGroupTypeDefault = "default"
	RuleGroupTypeDirect  = "direct"
)

type ServerFilter struct {
	Id     int64
	Tags   []string
	Group  int64
	Search string
	Page   int
	Size   int
}

// Deprecated: use internal/model/node/server.go
type Server struct {
	Id             int64     `gorm:"primary_key"`
	Name           string    `gorm:"type:varchar(100);not null;default:'';comment:Node Name"`
	Tags           string    `gorm:"type:varchar(128);not null;default:'';comment:Tags"`
	Country        string    `gorm:"type:varchar(128);not null;default:'';comment:Country"`
	City           string    `gorm:"type:varchar(128);not null;default:'';comment:City"`
	Latitude       string    `gorm:"type:varchar(128);not null;default:'';comment:Latitude"`
	Longitude      string    `gorm:"type:varchar(128);not null;default:'';comment:Longitude"`
	ServerAddr     string    `gorm:"type:varchar(100);not null;default:'';comment:Server Address"`
	RelayMode      string    `gorm:"type:varchar(20);not null;default:'none';comment:Relay Mode"`
	RelayNode      string    `gorm:"type:text;comment:Relay Node"`
	SpeedLimit     int       `gorm:"type:int;not null;default:0;comment:Speed Limit"`
	TrafficRatio   float32   `gorm:"type:DECIMAL(4,2);not null;default:0;comment:Traffic Ratio"`
	GroupId        int64     `gorm:"index:idx_group_id;type:int;default:null;comment:Group ID"`
	Protocol       string    `gorm:"type:varchar(20);not null;default:'';comment:Protocol"`
	Config         string    `gorm:"type:text;comment:Config"`
	Enable         *bool     `gorm:"type:tinyint(1);not null;default:1;comment:Enabled"`
	Sort           int64     `gorm:"type:int;not null;default:0;comment:Sort"`
	LastReportedAt time.Time `gorm:"comment:Last Reported Time"`
	CreatedAt      time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt      time.Time `gorm:"comment:Update Time"`
}

func (*Server) TableName() string {
	return "server"
}

func (s *Server) BeforeDelete(tx *gorm.DB) error {
	logger.Debugf("[Server] BeforeDelete")
	if err := tx.Exec("UPDATE `server` SET sort = sort - 1 WHERE sort > ?", s.Sort).Error; err != nil {
		return err
	}
	return nil
}

func (s *Server) BeforeUpdate(tx *gorm.DB) error {
	logger.Debugf("[Server] BeforeUpdate")
	var count int64
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&Server{}).
		Where("sort = ? AND id != ?", s.Sort, s.Id).Count(&count).Error; err != nil {
		return err
	}
	if count > 1 {
		// reorder sort
		if err := reorderSort(tx); err != nil {
			logger.Errorf("[Server] BeforeUpdate reorderSort error: %v", err.Error())
			return err
		}
		// get max sort
		var maxSort int64
		if err := tx.Model(&Server{}).Select("MAX(sort)").Scan(&maxSort).Error; err != nil {
			return err
		}
		s.Sort = maxSort + 1
	}
	return nil
}

func (s *Server) BeforeCreate(tx *gorm.DB) error {
	logger.Debugf("[Server] BeforeCreate")
	if s.Sort == 0 {
		var maxSort int64
		if err := tx.Model(&Server{}).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort).Error; err != nil {
			return err
		}
		s.Sort = maxSort + 1
	}
	return nil
}

type Vless struct {
	Port            int             `json:"port"`
	Flow            string          `json:"flow"`
	Transport       string          `json:"transport"`
	TransportConfig TransportConfig `json:"transport_config"`
	Security        string          `json:"security"`
	SecurityConfig  SecurityConfig  `json:"security_config"`
}

type Vmess struct {
	Port            int             `json:"port"`
	Flow            string          `json:"flow"`
	Transport       string          `json:"transport"`
	TransportConfig TransportConfig `json:"transport_config"`
	Security        string          `json:"security"`
	SecurityConfig  SecurityConfig  `json:"security_config"`
}

type Trojan struct {
	Port            int             `json:"port"`
	Flow            string          `json:"flow"`
	Transport       string          `json:"transport"`
	TransportConfig TransportConfig `json:"transport_config"`
	Security        string          `json:"security"`
	SecurityConfig  SecurityConfig  `json:"security_config"`
}

type Shadowsocks struct {
	Method    string `json:"method"`
	Port      int    `json:"port"`
	ServerKey string `json:"server_key"`
}

type Hysteria2 struct {
	Port           int            `json:"port"`
	HopPorts       string         `json:"hop_ports"`
	HopInterval    int            `json:"hop_interval"`
	ObfsPassword   string         `json:"obfs_password"`
	SecurityConfig SecurityConfig `json:"security_config"`
}

type Tuic struct {
	Port                 int            `json:"port"`
	DisableSNI           bool           `json:"disable_sni"`
	ReduceRtt            bool           `json:"reduce_rtt"`
	UDPRelayMode         string         `json:"udp_relay_mode"`
	CongestionController string         `json:"congestion_controller"`
	SecurityConfig       SecurityConfig `json:"security_config"`
}

type AnyTLS struct {
	Port           int            `json:"port"`
	SecurityConfig SecurityConfig `json:"security_config"`
}

type TransportConfig struct {
	Path        string `json:"path,omitempty"` // ws/httpupgrade
	Host        string `json:"host,omitempty"`
	ServiceName string `json:"service_name"` // grpc
}

type SecurityConfig struct {
	SNI               string `json:"sni"`
	AllowInsecure     bool   `json:"allow_insecure"`
	Fingerprint       string `json:"fingerprint"`
	RealityServerAddr string `json:"reality_server_addr"`
	RealityServerPort int    `json:"reality_server_port"`
	RealityPrivateKey string `json:"reality_private_key"`
	RealityPublicKey  string `json:"reality_public_key"`
	RealityShortId    string `json:"reality_short_id"`
}

type NodeRelay struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Prefix string `json:"prefix"`
}

type Group struct {
	Id          int64     `gorm:"primary_key"`
	Name        string    `gorm:"type:varchar(100);not null;default:'';comment:Group Name"`
	Description string    `gorm:"type:varchar(255);default:'';comment:Group Description"`
	CreatedAt   time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt   time.Time `gorm:"comment:Update Time"`
}

func (Group) TableName() string {
	return "server_group"
}

type RuleGroup struct {
	Id        int64     `gorm:"primary_key"`
	Icon      string    `gorm:"type:MEDIUMTEXT;comment:Rule Group Icon"`
	Name      string    `gorm:"type:varchar(100);not null;default:'';comment:Rule Group Name"`
	Type      string    `gorm:"type:varchar(100);not null;default:'';comment:Rule Group Type"`
	Tags      string    `gorm:"type:text;comment:Selected Node Tags"`
	Rules     string    `gorm:"type:MEDIUMTEXT;comment:Rules"`
	Enable    bool      `gorm:"type:tinyint(1);not null;default:1;comment:Rule Group Enable"`
	Default   bool      `gorm:"type:tinyint(1);not null;default:0;comment:Rule Group is Default"`
	CreatedAt time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt time.Time `gorm:"comment:Update Time"`
}

func (RuleGroup) TableName() string {
	return "server_rule_group"
}
func reorderSort(tx *gorm.DB) error {
	var servers []Server
	if err := tx.Order("sort, id").Find(&servers).Error; err != nil {
		return err
	}
	for i, server := range servers {
		if server.Sort != int64(i)+1 {
			if err := tx.Exec("UPDATE `server` SET sort = ? WHERE id = ?", i+1, server.Id).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
