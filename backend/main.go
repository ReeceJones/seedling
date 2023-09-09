package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"

	"github.com/golang-jwt/jwt"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	//go:embed config.json
	configData []byte
	config     Config
	jwtSecret  string = "secret"
	db         *gorm.DB
	port       int = *flag.Int("port", 8081, "port to listen on")
)

// config structs

type LinkConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type OCIConfig struct {
	ChartURL string `json:"chart_url"`
}

type RepoConfig struct {
	RepoURL   string `json:"repo_url"`
	ChartName string `json:"chart_name"`
}

type ValuePathConfig struct {
	Path string `json:"path"`
	Key  string `json:"key"`
}

type ValueConfig struct {
	Name        string            `json:"name"`
	Paths       []ValuePathConfig `json:"path"`
	Default     string            `json:"default"`
	Description string            `json:"description"`
	Manager     string            `json:"manager"`
}

type HelmConfig struct {
	RemoteType        string        `json:"remote_type"`
	ReleaseNameFormat string        `json:"release_name_format"`
	OCI               OCIConfig     `json:"oci"`
	Repo              RepoConfig    `json:"repo"`
	Values            []ValueConfig `json:"values"`
}

type ServiceConfig struct {
	Key         string       `json:"key"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Links       []LinkConfig `json:"links"`
	Icon        string       `json:"icon"`
	Tags        []string     `json:"tags"`
	Helm        HelmConfig   `json:"helm"`
}

type PortAllocatorConfig struct {
	StartPort uint16 `json:"start_port"`
	EndPort   uint16 `json:"end_port"`
}

type ManagersConfig struct {
	PortAllocator PortAllocatorConfig `json:"port_allocator"`
}

type Config struct {
	Managers ManagersConfig  `json:"managers"`
	Services []ServiceConfig `json:"services"`
}

// middleware

type AuthContext struct {
	echo.Context
	UserID   uint
	Username string
	Email    string
}

func AuthDecode(skipRoutes []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, skipRoute := range skipRoutes {
				if c.Path() == skipRoute {
					return next(c)
				}
			}
			authHeader := c.Request().Header.Get("Authorization")
			authComponents := strings.Split(authHeader, " ")
			if len(authComponents) != 2 {
				return c.JSON(http.StatusUnauthorized, "Invalid authorization header")
			}
			token := authComponents[1]
			jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil {
				return c.JSON(http.StatusUnauthorized, "Invalid token")
			}
			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok || !jwtToken.Valid {
				return c.JSON(http.StatusUnauthorized, "Invalid token")
			}
			userID := uint(claims["sub"].(float64))
			username := claims["username"].(string)
			email := claims["email"].(string)
			cc := &AuthContext{
				c,
				userID,
				username,
				email,
			}
			return next(cc)
		}
	}
}

// db models

type User struct {
	gorm.Model
	FirstName      string
	LastName       string
	Username       string
	Email          string
	HashedPassword string
}

type Service struct {
	gorm.Model
	Key         string
	Name        string
	Description string
	ProjectURL  string
}

type InstalledService struct {
	gorm.Model
	UserID    uint
	ServiceID uint
	User      User
	Service   Service
	LiveURL   string
}

type ServiceStatus struct {
	gorm.Model
	InstalledServiceID uint
	InstalledService   InstalledService
	Status             string
}

type ServiceVersion struct {
	gorm.Model
	ServiceID uint
	Service   Service
	Version   string
}

type Event struct {
	gorm.Model
	ServiceID   uint
	Service     Service
	Description string
}

type Filesystem struct {
	gorm.Model
	UserID uint
	User   User
	Name   string
	Path   string
}

type FilesystemAudit struct {
	gorm.Model
	FilesystemID  uint
	Filesystem    Filesystem
	BytesCapacity uint64
	BytesUsed     uint64
}

type NetworkAudit struct {
	gorm.Model
	ServiceID uint
	Service   Service
	UserID    uint
	User      User
	BytesIn   uint64
	BytesOut  uint64
}

// API schemas

type UserDataSchema struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type CreateUserRequestSchema struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type CreateUserResponseSchema struct {
	Ok   bool           `json:"ok"`
	Data UserDataSchema `json:"data"`
}

type LoginUserRequestSchema struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponseSchema struct {
	Ok    bool           `json:"ok"`
	Token string         `json:"token"`
	Data  UserDataSchema `json:"data"`
}

type ServiceDataSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	LiveURL     string `json:"project_url"`
}

type ServiceInfoSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ProjectURL  string `json:"project_url"`
	Links       []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"links"`
	Icon string   `json:"icon"`
	Tags []string `json:"tags"`
}

type ServiceListResponseSchema struct {
	Ok   bool                `json:"ok"`
	Data []ServiceDataSchema `json:"data"`
}

type ServiceGetInstalledResponseSchema struct {
	Ok   bool              `json:"ok"`
	Data ServiceDataSchema `json:"data"`
}

type ServiceGetInfoResponseSchema struct {
	Ok   bool              `json:"ok"`
	Data ServiceInfoSchema `json:"data"`
}

type ServiceInstallRequestSchema struct {
	Name string `json:"name"`
}

type ServiceInstallResponseSchema struct {
	Ok   bool              `json:"ok"`
	Data ServiceDataSchema `json:"data"`
}

type ServiceUnInstallResponseSchema struct {
	Ok bool `json:"ok"`
}

// database initializations

func initializeServiceSpecs() error {
	for _, service := range config.Services {
		dbService := Service{
			Key:         service.Key,
			Name:        service.Name,
			Description: service.Description,
		}
		tx := db.Where("key = ?", service.Key).First(&dbService)
		if tx.Error != nil {
			tx = db.Create(&dbService)
			if tx.Error != nil {
				return tx.Error
			}
		} else {
			dbService.Name = service.Name
			dbService.Description = service.Description
			tx = db.Save(&dbService)
			if tx.Error != nil {
				return tx.Error
			}
		}
	}

	return nil
}

func initializeDB() error {
	if err := initializeServiceSpecs(); err != nil {
		return err
	}

	return nil
}

// API handlers

func createUser(c echo.Context) error {
	var req CreateUserRequestSchema
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	tx := db.Where("username = ? or email = ?", req.Username, req.Email).First(&User{})
	if tx.Error == nil || tx.RowsAffected > 0 {
		return c.JSON(http.StatusBadRequest, "Username or email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		log.Printf("Failed to create password hash: %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	user := User{
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: string(hashedPassword),
	}
	tx = db.Save(&user)
	if tx.Error != nil {
		log.Printf("Failed to create user: %s", tx.Error)
		return c.JSON(http.StatusInternalServerError, tx.Error)
	}

	return c.JSON(http.StatusOK, CreateUserResponseSchema{
		Ok: true,
		Data: UserDataSchema{
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	})
}

func loginUser(c echo.Context) error {
	var req LoginUserRequestSchema
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	var user User
	tx := db.Where("email = ?", req.Email).First(&user)
	if tx.Error != nil {
		return c.JSON(http.StatusBadRequest, "Invalid email or password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid email or password")
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	claims["username"] = user.Username
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix()

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Failed to sign token: %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, LoginUserResponseSchema{
		Ok:    true,
		Token: tokenString,
		Data: UserDataSchema{
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	})
}

func getCurrentUser(c echo.Context) error {
	cc := c.(*AuthContext)
	userID := cc.UserID
	var user User
	tx := db.First(&user, userID)
	if tx.Error != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid token")
	}
	return c.JSON(http.StatusOK, UserDataSchema{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	})
}

func getInstalledServices(c echo.Context) error {
	cc := c.(*AuthContext)
	userID := cc.UserID
	var results []InstalledService
	tx := db.Preload("Service").Where("user_id = ?", userID).Find(&results)
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, tx.Error)
	}
	services := []ServiceDataSchema{}
	for _, installedService := range results {
		services = append(services, ServiceDataSchema{
			Name:        installedService.Service.Name,
			Description: installedService.Service.Description,
			LiveURL:     installedService.LiveURL,
		})
	}
	return c.JSON(http.StatusOK, ServiceListResponseSchema{
		Ok:   true,
		Data: services,
	})
}

func getInstalledService(c echo.Context) error {
	cc := c.(*AuthContext)
	userID := cc.UserID
	serviceName := c.Param("servicename")
	var installedService InstalledService
	tx := db.Preload("Service").Where("user_id = ? and name = ?", userID, serviceName).First(&installedService)
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, tx.Error)
	}
	return c.JSON(http.StatusOK, ServiceGetInstalledResponseSchema{
		Ok: true,
		Data: ServiceDataSchema{
			Name:        installedService.Service.Name,
			Description: installedService.Service.Description,
			LiveURL:     installedService.LiveURL,
		},
	})
}

func getServiceInfo(c echo.Context) error {
	serviceName := c.Param("servicename")
	var service Service
	tx := db.Where("name = ?", serviceName).First(&service)
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, tx.Error)
	}
	return c.JSON(http.StatusOK, ServiceGetInfoResponseSchema{
		Ok: true,
		Data: ServiceInfoSchema{
			Name:        service.Name,
			Description: service.Description,
			ProjectURL:  service.ProjectURL,
		},
	})
}

func installService(c echo.Context) error {
	cc := c.(*AuthContext)
	userID := cc.UserID
	var req ServiceInstallRequestSchema
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	log.Printf("Installing service %s for %s (%s, %d))\n", req.Name, cc.Username, cc.Email, cc.UserID)

	var service Service
	tx := db.Where("key = ?", req.Name).First(&service)
	if tx.Error != nil {
		return c.JSON(http.StatusBadRequest, "Invalid service name")
	}

	var installedService InstalledService
	tx = db.Where("user_id = ? and service_id = ?", userID, service.ID).First(&installedService)
	if tx.Error == nil {
		return c.JSON(http.StatusBadRequest, "Service already installed")
	}

	// get helm config for chart
	helmIdx := -1
	for i, s := range config.Services {
		if s.Key == service.Key {
			helmIdx = i
			break
		}
	}
	if helmIdx == -1 {
		return c.JSON(http.StatusInternalServerError, "Invalid service")
	}
	helmConfig := config.Services[helmIdx].Helm
	log.Printf("Using Helm Config: %+v\n", helmConfig)

	// get chart values
	valuesList := []string{}
	jsonValuesList := []string{}
	servicePort := uint16(0)
	for _, value := range helmConfig.Values {
		var valueStr string
		if value.Manager == "port_allocator" {
			port := config.Managers.PortAllocator.StartPort + uint16(installedService.ID)
			servicePort = port
			valueStr = fmt.Sprintf("%d", port)
		} else {
			valueStr = value.Default
		}

		for _, path := range value.Paths {
			if path.Key == "" {
				valuesList = append(valuesList, fmt.Sprintf("%s=%s", path, valueStr))
			} else {
				// construct JSON body
				jsonValue := fmt.Sprintf(`%s={"%s": "%s"}`, path.Path, path.Key, valueStr)
				jsonValuesList = append(jsonValuesList, jsonValue)
			}
		}
	}
	log.Printf("Using Chart Values: %+v\n", valuesList)

	// create helm client
	namespace := fmt.Sprintf("seedling-%s", cc.Username)
	helmClient, err := helmclient.New(&helmclient.Options{
		Namespace: namespace,
	})
	if err != nil {
		log.Printf("Failed to create helm client: %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	_, err = helmClient.InstallChart(context.Background(), &helmclient.ChartSpec{
		ReleaseName:     service.Key,
		ChartName:       helmConfig.OCI.ChartURL,
		Namespace:       namespace,
		Timeout:         5 * time.Minute,
		Wait:            true,
		UpgradeCRDs:     true,
		CreateNamespace: true,
		ValuesOptions: values.Options{ // causes hanging?
			Values:     valuesList,
			JSONValues: jsonValuesList,
		},
	}, nil)
	if err != nil {
		log.Printf("Failed to install chart: %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	installedService = InstalledService{
		UserID:    userID,
		ServiceID: service.ID,
	}
	tx = db.Create(&installedService)
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, tx.Error)
	}

	return c.JSON(http.StatusOK, ServiceInstallResponseSchema{
		Ok: true,
		Data: ServiceDataSchema{
			Name:        service.Name,
			Description: service.Description,
			LiveURL:     fmt.Sprintf(":%d", servicePort),
		},
	})
}

func uninstallService(c echo.Context) error {
	cc := c.(*AuthContext)
	userID := cc.UserID
	serviceName := c.Param("servicename")
	var installedService InstalledService
	tx := db.Table("installed_services").Joins("JOIN services ON installed_services.service_id = services.id").Where("installed_services.user_id = ? and services.key = ?", userID, serviceName).Select("installed_services.*").First(&installedService)
	// tx := db.Preload("Service").Where("user_id = ? and name = ?", userID, serviceName).First(&installedService)
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, tx.Error)
	}

	// create helm client
	namespace := fmt.Sprintf("seedling-%s", cc.Username)
	helmClient, err := helmclient.New(&helmclient.Options{
		Namespace: namespace,
	})
	if err != nil {
		log.Printf("Failed to create helm client: %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = helmClient.UninstallRelease(&helmclient.ChartSpec{
		ReleaseName: serviceName,
		Namespace:   namespace,
	})
	if err != nil {
		log.Printf("Failed to uninstall chart: %s", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	tx = db.Delete(&installedService)
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, tx.Error)
	}

	return c.JSON(http.StatusOK, ServiceUnInstallResponseSchema{
		Ok: true,
	})
}

func main() {
	// load config
	err := json.Unmarshal(configData, &config)
	if err != nil {
		log.Fatal(err)
	}

	// configure echo
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(jwtSecret),
		Skipper: func(c echo.Context) bool {
			if ((c.Path() == "" || c.Path() == "/") && c.Request().Method == "GET") || ((c.Path() == "/v1/users" || c.Path() == "/v1/users/login") && c.Request().Method == "POST") {
				return true
			}
			return false
		},
	}))
	e.Use(AuthDecode([]string{"", "/", "/v1/users", "/v1/users/login"}))

	// create routes
	e.GET("/", func(c echo.Context) error { return c.String(http.StatusOK, "") })

	v1Group := e.Group("/v1")

	usersGroup := v1Group.Group("/users")

	usersGroup.POST("", createUser).Name = "create-user"
	usersGroup.POST("/login", loginUser).Name = "login"
	usersGroup.GET("", getCurrentUser).Name = "get-current-user"
	// usersGroup.PATCH("", updateUser).Name = "update-current-user"
	// usersGroup.DELETE("", deleteUser).Name = "delete-current-user"

	servicesGroup := v1Group.Group("/services")
	servicesGroup.GET("/installed", getInstalledServices).Name = "get-installed-services"
	servicesGroup.GET("/installed/:servicename", getInstalledService).Name = "get-installed-service"
	servicesGroup.GET("/info/:servicename", getServiceInfo).Name = "get-service-info"
	servicesGroup.POST("/install", installService).Name = "install-service"
	servicesGroup.POST("/uninstall/:servicename", uninstallService).Name = "uninstall-service"

	// eventsGroup := v1Group.Group("/events")
	// eventsGroup.GET("", getEvents).Name = "get-events"
	// eventsGroup.GET("/:eventid", getEvent).Name = "get-event"

	// filesystemsGroup := v1Group.Group("/filesystems")
	// filesystemsGroup.GET("", getFilesystems).Name = "get-filesystems"
	// filesystemsGroup.GET("/:name", getFilesystem).Name = "get-filesystem"
	// filesystemsGroup.POST("", createFilesystem).Name = "create-filesystem"
	// filesystemsGroup.PATCH("/:name", updateFilesystem).Name = "update-filesystem"
	// filesystemsGroup.DELETE("/:name", deleteFilesystem).Name = "delete-filesystem"

	// networkGroup := v1Group.Group("/network")
	// networkGroup.GET("", getNetworkUsage).Name = "get-network-usage"

	// open database connection
	sqliteDB, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db = sqliteDB
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Service{})
	db.AutoMigrate(&InstalledService{})
	db.AutoMigrate(&ServiceStatus{})
	db.AutoMigrate(&ServiceVersion{})
	db.AutoMigrate(&Event{})
	db.AutoMigrate(&Filesystem{})
	db.AutoMigrate(&FilesystemAudit{})
	db.AutoMigrate(&NetworkAudit{})

	// run final database initializations
	initializeDB()

	// start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}
