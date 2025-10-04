package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository() (*Repository, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &Repository{db: db}

	// Инициализация таблиц
	if err := repo.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return repo, nil
}

func (r *Repository) GetDB() *sql.DB {
	return r.db
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) initTables() error {
	// Если управляем миграциями через Prisma, можно пропустить авто-создание
	if os.Getenv("PRISMA_MANAGED") == "1" {
		return nil
	}
	// Новая схема пользователей с bcrypt + солью
	usersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        username VARCHAR(64) UNIQUE NOT NULL,
        email VARCHAR(128) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        password_salt VARCHAR(64) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	if _, err := r.db.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Таблица servers (если ещё не создана вручную)
	serversTable := `
    CREATE TABLE IF NOT EXISTS servers (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(128) NOT NULL,
        host VARCHAR(255) NOT NULL,
        port INT NOT NULL DEFAULT 0,
        type CHAR(3) NOT NULL,
        username_enc TEXT NOT NULL,
        password_enc TEXT NOT NULL,
        user_id INT NOT NULL,
        is_active TINYINT(1) NOT NULL DEFAULT 1,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
        INDEX (user_id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	if _, err := r.db.Exec(serversTable); err != nil {
		return fmt.Errorf("failed to create servers table: %w", err)
	}

	// Если старое поле password осталось (миграция не выполнена) — попытаться переименовать (best effort)
	_, _ = r.db.Exec("ALTER TABLE users CHANGE COLUMN password password_hash VARCHAR(255)")
	// Если не хватает столбца password_salt — добавить
	_, _ = r.db.Exec("ALTER TABLE users ADD COLUMN password_salt VARCHAR(64) NOT NULL AFTER password_hash")

	return nil
}
