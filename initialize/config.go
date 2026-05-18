package initialize

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/perfect-panel/server/internal/report"
	"github.com/perfect-panel/server/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/perfect-panel/server/initialize/migrate"
	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/pkg/conf"
	"github.com/perfect-panel/server/pkg/orm"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
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
	// get server port
	port := 8080
	host := "127.0.0.1"

	// check gateway mode
	if report.IsGatewayMode() {
		// get free port
		freePort, err := report.ModulePort()
		if err != nil {
			logger.Errorf("get module port error: %s", err.Error())
			panic(err)
		}
		port = freePort
		// register module
		err = report.RegisterModule(port)
		if err != nil {
			logger.Errorf("register module error: %s", err.Error())
			panic(err)
		}
		logger.Infof("module registered on port %d", port)
	}
	// Create a new HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: r,
	}
	// Load templates
	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/init", handleInit)
	r.POST("/init/config", handleInitConfig)
	r.POST("/init/database/test", HandleDatabaseTest)
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

		DatabaseDriver string `json:"databaseDriver"`
		MysqlHost      string `json:"mysqlHost"`
		MysqlPort      string `json:"mysqlPort"`
		MysqlDatabase  string `json:"mysqlDatabase"`
		MysqlUser      string `json:"mysqlUser"`
		MysqlPassword  string `json:"mysqlPassword"`

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
	// database
	dbConfig, err := buildDatabaseConfig(request.DatabaseDriver, request.MysqlHost, request.MysqlPort, request.MysqlDatabase, request.MysqlUser, request.MysqlPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
			"data": nil,
		})
		c.Abort()
		return
	}
	cfg.SetDatabaseConfig(dbConfig)
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

	// create database connection
	dbClient := orm.Mysql{Config: dbConfig}
	db, err := orm.ConnectDatabase(dbClient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Database connection failed",
			"data": nil,
		})
		c.Abort()
		return
	}
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}
	// migrate database
	if err = migrate.Migrate(dbClient.Driver(), dbClient.MigrationDsn()).Up(); err != nil {
		logger.Errorf("[Init Database] Migrate failed: %v", err.Error())
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
		logger.Errorf("[Init Database] Create admin user failed: %v", err.Error())
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
	HandleDatabaseTest(c)
}

func HandleDatabaseTest(c *gin.Context) {
	var request struct {
		Driver   string `json:"driver"`
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
	var status = true
	var message string
	var tx *sql.DB
	var tables []string
	dbConfig, err := buildDatabaseConfig(request.Driver, request.Host, request.Port, request.Database, request.User, request.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":   200,
			"msg":    err.Error(),
			"status": false,
		})
		return
	}
	db, err := orm.ConnectDatabase(orm.Mysql{Config: dbConfig})
	if err != nil {
		logger.Errorf("connect database failed, err: %v\n", err.Error())
		status = false
		message = "Database connection failed"
		goto result
	}
	tx, _ = db.DB()
	if tx != nil {
		defer tx.Close()
	}
	if err := tx.Ping(); err != nil {
		logger.Errorf("ping database failed, err: %v\n", err.Error())
		status = false
		message = "Database connection failed"
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

func buildDatabaseConfig(driver, host, port, database, user, password string) (orm.Config, error) {
	normalizedDriver := orm.NormalizeDriver(driver)
	switch normalizedDriver {
	case orm.DriverMySQL, orm.DriverPostgres:
	default:
		return orm.Config{}, fmt.Errorf("unsupported database driver: %s", driver)
	}
	cfg := orm.Config{
		Driver:        normalizedDriver,
		Addr:          fmt.Sprintf("%s:%s", host, port),
		Username:      user,
		Password:      password,
		Dbname:        database,
		MaxIdleConns:  10,
		MaxOpenConns:  10,
		SlowThreshold: orm.DefaultSlowThresholdMs,
	}
	if normalizedDriver == orm.DriverPostgres {
		cfg.Config = orm.DefaultPostgresConfig
	} else {
		cfg.Config = orm.DefaultMySQLConfig
	}
	return cfg, nil
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
