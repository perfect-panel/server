package initialize

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/perfect-panel/server/pkg/logger"
	"gorm.io/driver/mysql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/perfect-panel/server/initialize/migrate"
	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/pkg/conf"
	"github.com/perfect-panel/server/pkg/orm"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

//go:embed templates/*.html
var templateFS embed.FS

var initStatus = make(chan bool)
var configPath string

func Config(path string) (chan bool, *http.Server) {
	// Set the configuration file path
	configPath = path
	// Create a new Gin instance
	r := gin.Default()

	// Create a new HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	// Load templates
	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/init", handleInit)
	r.POST("/init/config", handleInitConfig)
	r.POST("/init/mysql/test", HandleMySQLTest)
	r.POST("/init/redis/test", HandleRedisTest)
	// Handle 404
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/init")
	})

	go func(server *http.Server) {
		// Start the server
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}(server)

	return initStatus, server
}

func handleInit(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}
func handleInitConfig(c *gin.Context) {
	// Load configuration file

	var cfg config.File
	conf.MustLoad(configPath, &cfg)
	var request struct {
		AdminEmail    string `json:"adminEmail"`
		AdminPassword string `json:"adminPassword"`

		MysqlHost     string `json:"mysqlHost"`
		MysqlPort     string `json:"mysqlPort"`
		MysqlDatabase string `json:"mysqlDatabase"`
		MysqlUser     string `json:"mysqlUser"`
		MysqlPassword string `json:"mysqlPassword"`

		RedisHost     string `json:"redisHost"`
		RedisPort     string `json:"redisPort"`
		RedisPassword string `json:"redisPassword"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request",
			"data": nil,
		})
		c.Abort()
		return
	}
	cfg.Debug = false
	// jwt secret
	cfg.JwtAuth.AccessSecret = uuid.New().String()
	// mysql
	cfg.MySQL.Addr = fmt.Sprintf("%s:%s", request.MysqlHost, request.MysqlPort)
	cfg.MySQL.Dbname = request.MysqlDatabase
	cfg.MySQL.Username = request.MysqlUser
	cfg.MySQL.Password = request.MysqlPassword
	// redis
	cfg.Redis.Host = fmt.Sprintf("%s:%s", request.RedisHost, request.RedisPort)
	cfg.Redis.Pass = request.RedisPassword

	// save config
	fileData, err := yaml.Marshal(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Configuration initialization failed",
			"data": nil,
		})
		c.Abort()
		return
	}

	// create mysql connection
	db, err := orm.ConnectMysql(orm.Mysql{
		Config: orm.Config{
			Addr:          fmt.Sprintf("%s:%s", request.MysqlHost, request.MysqlPort),
			Username:      request.MysqlUser,
			Password:      request.MysqlPassword,
			Dbname:        request.MysqlDatabase,
			Config:        "charset%3Dutf8mb4%26parseTime%3Dtrue%26loc%3DLocal",
			MaxIdleConns:  10,
			MaxOpenConns:  10,
			SlowThreshold: 1000,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "MySQL connection failed",
			"data": nil,
		})
		c.Abort()
		return
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", request.MysqlUser, request.MysqlPassword, request.MysqlHost, request.MysqlPort, request.MysqlDatabase)
	// migrate database
	if err = migrate.Migrate(dsn).Up(); err != nil {
		logger.Errorf("[Init Mysql] Migrate failed: %v", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "Database migration failed",
			"data": nil,
		})
		c.Abort()
		return
	}

	// create admin user
	if err = migrate.CreateAdminUser(request.AdminEmail, request.AdminPassword, db); err != nil {
		logger.Errorf("[Init Mysql] Create admin user failed: %v", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "Admin user creation failed",
			"data": nil,
		})
		c.Abort()
		return
	}

	// write to file
	if err = os.WriteFile(configPath, fileData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Configuration initialization failed",
			"data": nil,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"msg":    "Configuration initialized",
		"status": true,
	})
	initStatus <- true
}

func HandleMySQLTest(c *gin.Context) {
	var request struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Database string `json:"database"`
		User     string `json:"user"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request",
			"data": nil,
		})
		c.Abort()
		return
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", request.User, request.Password, request.Host, request.Port, request.Database)
	var status = true
	var message string
	var tx *sql.DB
	var tables []string
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Errorf("connect mysql failed, err: %v\n", err.Error())
		status = false
		message = "MySQL connection failed"
		goto result
	}
	tx, _ = db.DB()
	if err := tx.Ping(); err != nil {
		logger.Errorf("ping mysql failed, err: %v\n", err.Error())
		status = false
		message = "MySQL connection failed"
	}

	tables, err = db.Migrator().GetTables()
	if err != nil {
		logger.Errorf("database table check failed, err: %v\n", err.Error())
		status = false
		message = "Database table check failed"
		goto result
	}
	if len(tables) > 0 {
		status = false
		message = "The database contains existing data. Please clear it before proceeding with the installation."
		goto result
	}

result:
	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"msg":    message,
		"status": status,
	})
}

func HandleRedisTest(c *gin.Context) {
	var request struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request",
			"data": nil,
		})
		c.Abort()
		return
	}
	if err := tool.RedisPing(fmt.Sprintf("%s:%s", request.Host, request.Port), request.Password, 0); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":   200,
			"msg":    nil,
			"status": false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"msg":    nil,
		"status": true,
	})
}
