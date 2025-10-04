package server

import "time"

type Server struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Type        string `json:"type"` // prx, vmv, hyv, kvm, xen
	UsernameEnc string `json:"-"`
	PasswordEnc string `json:"-"`
	// Дешифрованные значения (runtime)
	UsernameDecrypted string    `json:"username"`
	PasswordDecrypted string    `json:"-"`
	UserID            int       `json:"user_id"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreateServerRequest struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateServerRequest struct {
	Name     string `json:"name,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}
