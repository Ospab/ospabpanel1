package user

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"

	mysql "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db *sql.DB
}

// Sentinel errors for business logic / presentation layer mapping
var (
	ErrUsernameTaken    = errors.New("username_taken")
	ErrEmailTaken       = errors.New("email_taken")
	ErrDuplicateValue   = errors.New("duplicate_value")
	ErrPasswordTooShort = errors.New("password_too_short")
)

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := "SELECT id, username, email, password_hash, password_salt, created_at, updated_at FROM users WHERE username = ?"
	err := s.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.PasswordSalt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) GetUserByID(id int) (*User, error) {
	user := &User{}
	query := "SELECT id, username, email, password_hash, password_salt, created_at, updated_at FROM users WHERE id = ?"
	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.PasswordSalt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func generateSalt(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(b), nil
}

func hashPassword(password, salt string) (string, error) {
	// Добавляем соль в пароль (простой конкатенационный подход, bcrypt сам добавляет свою соль
	// но мы усиливаем связку пользовательской солью)
	combined := password + ":" + salt
	h, err := bcrypt.GenerateFromPassword([]byte(combined), bcrypt.DefaultCost)
	return string(h), err
}

func (s *Service) ValidatePassword(u *User, password string) bool {
	combined := password + ":" + u.PasswordSalt
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(combined)) == nil
}

func (s *Service) CreateUser(username, email, password string) (*User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))
	if username == "" || email == "" || password == "" {
		return nil, errors.New("invalid empty fields")
	}
	if len(password) < 8 { // минимальная длина
		return nil, ErrPasswordTooShort
	}

	salt, err := generateSalt(16)
	if err != nil {
		return nil, err
	}
	hash, err := hashPassword(password, salt)
	if err != nil {
		return nil, err
	}

	query := "INSERT INTO users (username, email, password_hash, password_salt, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())"
	result, err := s.db.Exec(query, username, email, hash, salt)
	if err != nil {
		// Обработка дублей (email/username)
		if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1062 {
			lerr := strings.ToLower(me.Message)
			if strings.Contains(lerr, "username") {
				return nil, ErrUsernameTaken
			}
			if strings.Contains(lerr, "email") {
				return nil, ErrEmailTaken
			}
			return nil, ErrDuplicateValue
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(int(id))
}
