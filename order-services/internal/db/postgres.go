package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"order-services/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	defaultMigrationsDir = "migrations"
	migrationsTableName  = "schema_migrations"
)

type migrationFile struct {
	version    string
	versionNum int
	name       string
	path       string
}

// ConnectPostgres creates a GORM connection to PostgreSQL
func ConnectPostgres() (*gorm.DB, error) {
	cfg := config.GetConfig()
	dsn := cfg.DatabaseURL()

	// Configure GORM logger based on environment
	logLevel := logger.Silent
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL with GORM")
	return db, nil
}

// RunMigrations executes SQL files stored in the migrations directory.
// Files should follow the pattern: <version>_<description>.up.sql (e.g. 0001_initial.up.sql).
func RunMigrations(gormDB *gorm.DB, dirOverride string) error {
	if err := ensureUUIDExtension(gormDB); err != nil {
		return err
	}

	dir := resolveMigrationsDir(dirOverride)
	migrations, err := loadMigrationFiles(dir)
	if err != nil {
		return err
	}
	if len(migrations) == 0 {
		log.Printf("⚠️  No migration files found in %s", dir)
		return nil
	}

	if err := ensureMigrationsTable(gormDB); err != nil {
		return err
	}

	appliedCount := 0
	for _, m := range migrations {
		applied, err := isMigrationApplied(gormDB, m.version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigrationFile(gormDB, m); err != nil {
			return err
		}
		log.Printf("✅ Applied migration %s", m.name)
		appliedCount++
	}

	if appliedCount == 0 {
		log.Println("✅ Database already up to date")
	} else {
		log.Printf("✅ Completed %d migration(s)", appliedCount)
	}
	return nil
}

func resolveMigrationsDir(dir string) string {
	if dir != "" {
		return dir
	}
	if envDir := os.Getenv("MIGRATIONS_DIR"); envDir != "" {
		return envDir
	}
	return defaultMigrationsDir
}

func ensureUUIDExtension(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf("failed to enable uuid-ossp extension: %w", err)
	}
	return nil
}

func ensureMigrationsTable(db *gorm.DB) error {
	createStmt := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`, migrationsTableName)
	if err := db.Exec(createStmt).Error; err != nil {
		return fmt.Errorf("failed to create %s table: %w", migrationsTableName, err)
	}
	return nil
}

func loadMigrationFiles(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory %s: %w", dir, err)
	}

	files := make([]migrationFile, 0, len(entries))
	seen := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".sql") {
			continue
		}
		if strings.Contains(lower, ".down.") {
			continue
		}
		version, versionNum := parseMigrationVersion(name)
		if version == "" {
			log.Printf("⚠️  Skipping migration with unrecognized name: %s", name)
			continue
		}
		if _, exists := seen[version]; exists {
			return nil, fmt.Errorf("duplicate migration version detected: %s", version)
		}
		seen[version] = struct{}{}
		files = append(files, migrationFile{
			version:    version,
			versionNum: versionNum,
			name:       name,
			path:       filepath.Join(dir, name),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].versionNum == files[j].versionNum {
			return files[i].name < files[j].name
		}
		return files[i].versionNum < files[j].versionNum
	})

	return files, nil
}

func parseMigrationVersion(name string) (string, int) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", 0
	}
	parts := strings.SplitN(name, "_", 2)
	raw := parts[0]
	numStr := strings.TrimLeft(raw, "0")
	if numStr == "" {
		numStr = "0"
	}
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return "", 0
	}
	return raw, num
}

func isMigrationApplied(db *gorm.DB, version string) (bool, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE version = ?", migrationsTableName)
	if err := db.Raw(query, version).Scan(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check migration %s: %w", version, err)
	}
	return count > 0, nil
}

func applyMigrationFile(db *gorm.DB, m migrationFile) error {
	content, err := os.ReadFile(m.path)
	if err != nil {
		return fmt.Errorf("failed to read migration %s: %w", m.name, err)
	}

	sql := strings.TrimSpace(string(content))
	if sql == "" {
		log.Printf("⚠️  Migration %s is empty, marking as applied", m.name)
		return recordMigration(db, m.version)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(sql).Error; err != nil {
			return fmt.Errorf("migration %s failed: %w", m.name, err)
		}
		if err := recordMigration(tx, m.version); err != nil {
			return err
		}
		return nil
	})
}

func recordMigration(db *gorm.DB, version string) error {
	stmt := fmt.Sprintf("INSERT INTO %s (version) VALUES (?)", migrationsTableName)
	if err := db.Exec(stmt, version).Error; err != nil {
		return fmt.Errorf("failed to record migration %s: %w", version, err)
	}
	return nil
}
