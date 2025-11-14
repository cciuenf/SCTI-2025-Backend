package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	mp_config "github.com/mercadopago/sdk-go/pkg/config"
)

type Config struct {
	DB         string
	DB_NAME    string
	DB_PASS    string
	DB_PORT    string
	DB_USER    string
	DSN        string
	HOST       string
	PORT       string
	JWT_SECRET string
}

var (
	server_host            string
	server_port            string
	db                     string
	db_port                string
	db_user                string
	db_pass                string
	jwtSecret              string
	dsn                    string
	systemEmail            string
	emailPass              string
	masterUserPass         string
	siteURL                string
	mercadoPagoAccessToken string
	mercadoPagoPublicKey   string
	mercadoPagoConfig      *mp_config.Config
	webhook_signature      string
)

func LoadConfig(path string) *Config {
	err := godotenv.Load(path)
	if err != nil {
		log.Printf("Could not load %s file, using environment variables: %v", path, err)
	}

	server_host = os.Getenv("HOST")
	server_port = os.Getenv("PORT")
	db = os.Getenv("DATABASE")
	db_port = os.Getenv("DATABASE_PORT")
	db_user = os.Getenv("DATABASE_USER")
	db_pass = os.Getenv("DATABASE_PASS")
	db_host := os.Getenv("DB_HOST")
	jwtSecret = os.Getenv("JWT_SECRET")
	systemEmail = os.Getenv("SCTI_EMAIL")
	masterUserPass = os.Getenv("MASTER_USER_PASS")
	emailPass = os.Getenv("SCTI_APP_PASSWORD")
	siteURL = os.Getenv("SITE_URL")
	mercadoPagoAccessToken = os.Getenv("MERCADO_PAGO_ACCESS_TOKEN")
	mercadoPagoPublicKey = os.Getenv("MERCADO_PAGO_PUBLIC_KEY")
	webhook_signature = os.Getenv("WEBHOOK_SIGNATURE")

	accessToken := mercadoPagoAccessToken
	mercadoPagoConfig, err = mp_config.New(accessToken)
	if err != nil {
		log.Fatalf("Failed to create mercado pago config: %v", err)
	}

	dsn = fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=America/Sao_Paulo", db_host, db_user, db_pass, db, db_port)

	return &Config{
		HOST:       server_host,
		PORT:       server_port,
		DB:         db,
		DB_PORT:    db_port,
		DB_USER:    db_user,
		DB_PASS:    db_pass,
		DSN:        dsn,
		JWT_SECRET: jwtSecret,
	}
}

func GetServerHost() string {
	return server_host
}

func GetServerPort() string {
	return server_port
}

func GetDB() string {
	return db
}

func GetDBPort() string {
	return db_port
}

func GetDBUser() string {
	return db_user
}

func GetDBPass() string {
	return db_pass
}

func GetJWTSecret() string {
	return jwtSecret
}

func GetDSN() string {
	return dsn
}

func GetSystemEmail() string {
	return systemEmail
}

func GetSystemEmailPass() string {
	return emailPass
}

func GetMasterUserPass() string {
	return masterUserPass
}

func GetSiteURL() string {
	return siteURL
}

func GetMercadoPagoAccessToken() string {
	return mercadoPagoAccessToken
}

func GetMercadoPagoPublicKey() string {
	return mercadoPagoPublicKey
}

func GetMercadoPagoConfig() *mp_config.Config {
	return mercadoPagoConfig
}

func GetWebhookSignature() string {
	return webhook_signature
}
