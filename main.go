package main

import (
	"database/sql"
	"fmt"
	_ "github.com/Fyefhqdishka/song-library/docs"
	"github.com/Fyefhqdishka/song-library/routes"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pressly/goose"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	logger := initLogging()
	logger.Info("Приложение запущено")

	if err := loadEnv(logger); err != nil {
		logger.Error("Ошибка загрузки окружения", "error", err)
		os.Exit(1)
	}

	logger.Info("Окружение загружено")

	db, err := connectToDB()
	if err != nil {
		logger.Error("Ошибка подключения к базе данных", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	logger.Info("Подключение к базе данных успешно")

	if err = goose.Up(db, "./migrations"); err != nil {
		logger.Error("Ошибка применения миграций", "error", err)
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		logger.Error("Ошибка пинга базы данных", "error", err)
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	routes.RegisterRoutes(r, db, logger)

	port := ":3000"
	if err := http.ListenAndServe(port, r); err != nil {
		logger.Error("Error starting server: %v", err)
		os.Exit(1)
	}
}

func loadEnv(logger *slog.Logger) error {
	err := godotenv.Load("/app/.env")
	if err != nil {
		logger.Error("Ошибка загрузки окружения", "error", err)
		return err
	}
	return nil
}

func initLogging() *slog.Logger {
	logFileName := "logs/app-" + time.Now().Format("2006-01-02") + ".log"
	logfile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("Не удалось открыть файл для логов", "error", err)
		os.Exit(1)
	}

	handler := slog.NewTextHandler(logfile, nil)
	return slog.New(handler)
}

func connectToDB() (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)
	return sql.Open("postgres", connStr)
}
