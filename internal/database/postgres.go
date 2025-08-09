package database

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"smat/iot/simulation/iot-inventory-management/internal/config"
)

func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// RunMigrations runs all pending migrations
func RunMigrations(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	log.Printf("Running migrations from directory: %s", absPath)

	if err := goose.Up(db, absPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	newVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get version after migration: %w", err)
	}

	log.Printf("Migrations completed successfully. Current version: %d", newVersion)
	return nil
}

// GetRequiredClientIDs returns the list of client IDs that are needed for the application
func GetRequiredClientIDs() []struct {
	ID       string
	Name     string
	Location string
} {
	return []struct {
		ID       string
		Name     string
		Location string
	}{
		{"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", "Client A", "New York"},
		{"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", "Client B", "Los Angeles"},
		{"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13", "Client C", "Chicago"},
		{"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14", "Client D", "Houston"},
		{"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a15", "Client E", "Phoenix"},
	}
}

// CheckAndFixClientData ensures all required clients exist
func CheckAndFixClientData(db *sql.DB) error {
	log.Println("Checking client data integrity...")

	requiredClients := GetRequiredClientIDs()

	// Check which clients exist
	existingClients := make(map[string]bool)
	rows, err := db.Query("SELECT id FROM clients")
	if err != nil {
		return fmt.Errorf("failed to query existing clients: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan client id: %w", err)
		}
		existingClients[id] = true
	}

	log.Printf("Found %d existing clients in database", len(existingClients))

	// Check which required clients are missing
	var missingClients []struct {
		ID       string
		Name     string
		Location string
	}

	for _, client := range requiredClients {
		if !existingClients[client.ID] {
			missingClients = append(missingClients, client)
			log.Printf("Missing required client: %s (%s)", client.Name, client.ID)
		} else {
			log.Printf("Found required client: %s (%s)", client.Name, client.ID)
		}
	}

	// Insert missing clients
	if len(missingClients) > 0 {
		log.Printf("Inserting %d missing clients...", len(missingClients))

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare("INSERT INTO clients (id, name, location) VALUES ($1, $2, $3)")
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, client := range missingClients {
			_, err := stmt.Exec(client.ID, client.Name, client.Location)
			if err != nil {
				return fmt.Errorf("failed to insert client %s: %w", client.Name, err)
			}
			log.Printf("Inserted missing client: %s (%s)", client.Name, client.ID)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		log.Printf("Successfully inserted %d missing clients", len(missingClients))
	} else {
		log.Println("All required clients exist")
	}

	return nil
}

// SeedDevices inserts sample devices if they don't exist
func SeedDevices(db *sql.DB) error {
	log.Println("Checking device data...")

	// Check if devices already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM devices").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing devices: %w", err)
	}

	if count > 0 {
		log.Printf("Found %d existing devices, skipping device seed", count)
		return nil
	}

	log.Println("No devices found, seeding sample devices...")

	// Verify all required clients exist before inserting devices
	requiredClients := GetRequiredClientIDs()
	for _, client := range requiredClients {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)", client.ID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check client existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("required client %s does not exist, cannot seed devices", client.ID)
		}
	}

	// Insert sample devices
	devices := []struct {
		DeviceID    string
		ClientID    string
		Location    string
		MaxCapacity float64
		ItemWeight  float64
	}{
		{"DEV001", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", "Warehouse A1", 500.00, 2.50},
		{"DEV002", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", "Warehouse A2", 750.00, 3.00},
		{"DEV003", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", "Storage B1", 300.00, 1.75},
		{"DEV004", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13", "Facility C1", 600.00, 2.25},
		{"DEV005", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14", "Hub D1", 400.00, 2.00},
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
        INSERT INTO devices (device_id, client_id, location, max_capacity, item_weight) 
        VALUES ($1, $2, $3, $4, $5)
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, device := range devices {
		_, err := stmt.Exec(device.DeviceID, device.ClientID, device.Location,
			device.MaxCapacity, device.ItemWeight)
		if err != nil {
			return fmt.Errorf("failed to insert device %s: %w", device.DeviceID, err)
		}
		log.Printf("Inserted device: %s for client %s", device.DeviceID, device.ClientID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully seeded %d devices", len(devices))
	return nil
}

// InitializeDatabase runs migrations and seeds initial data
func InitializeDatabase(db *sql.DB) error {
	// Run migrations first
	if err := RunMigrations(db, "./migrations"); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Check and fix client data (ensures all required clients exist)
	if err := CheckAndFixClientData(db); err != nil {
		return fmt.Errorf("client data check failed: %w", err)
	}

	// Then seed devices
	if err := SeedDevices(db); err != nil {
		return fmt.Errorf("device seeding failed: %w", err)
	}

	return nil
}

// CleanDatabase removes all data (useful for development/testing)
func CleanDatabase(db *sql.DB) error {
	log.Println("Cleaning database...")

	tables := []string{"inventory_readings", "devices", "clients"}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clean table %s: %w", table, err)
		}
		log.Printf("Cleaned table: %s", table)
	}

	log.Println("Database cleaned successfully")
	return nil
}
