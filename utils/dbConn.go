package utils

import (
	"database/sql"

	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func NewDatabaseConnection() (*sql.DB, error) {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	if user == "" || password == "" || dbName == "" {
		return nil, fmt.Errorf("missing required database env vars (POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}

func GetCameraData() ([]Camera, error) {
	db, err := NewDatabaseConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get camera data: %w", err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM cameras")
	if err != nil {
		return nil, fmt.Errorf("failed to get camera data: %w", err)
	}
	defer rows.Close()

	results := []Camera{}
	for rows.Next() {
		var camera Camera
		err := rows.Scan(&camera.ID, &camera.Name, &camera.Rtsp_url)
		if err != nil {
			return nil, fmt.Errorf("failed to scan camera data: %w", err)
		}
		results = append(results, camera)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %w", err)
	}
	return results, nil
}
