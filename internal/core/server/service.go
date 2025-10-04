package server

import (
	"database/sql"
	"fmt"
	"os"
	"ospab-panel/internal/infra/crypto"
	"strings"
)

type Service struct{ db *sql.DB }

func (s *Service) key() ([]byte, error) {
	k := os.Getenv("SERVER_SECRET_KEY")
	if len(k) == 0 { // fallback dev insecure
		k = "dev-insecure-key-dev-insecure-key-32!!" // 32 bytes
	}
	if len(k) < 32 {
		k = (k + strings.Repeat("0", 32))[:32]
	}
	return []byte(k[:32]), nil
}

func NewService(db *sql.DB) *Service { return &Service{db: db} }

func (s *Service) GetServersByUserID(userID int) ([]*Server, error) {
	rows, err := s.db.Query(`SELECT id,name,host,port,type,username_enc,password_enc,user_id,is_active,created_at,updated_at FROM servers WHERE user_id = ? AND is_active=1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*Server{}
	for rows.Next() {
		var srv Server
		if err := rows.Scan(&srv.ID, &srv.Name, &srv.Host, &srv.Port, &srv.Type, &srv.UsernameEnc, &srv.PasswordEnc, &srv.UserID, &srv.IsActive, &srv.CreatedAt, &srv.UpdatedAt); err != nil {
			return nil, err
		}
		if err := s.decryptRuntime(&srv); err != nil {
			return nil, err
		}
		list = append(list, &srv)
	}
	return list, nil
}

func (s *Service) GetServerByID(id, userID int) (*Server, error) {
	var srv Server
	err := s.db.QueryRow(`SELECT id,name,host,port,type,username_enc,password_enc,user_id,is_active,created_at,updated_at FROM servers WHERE id=? AND user_id=?`, id, userID).Scan(&srv.ID, &srv.Name, &srv.Host, &srv.Port, &srv.Type, &srv.UsernameEnc, &srv.PasswordEnc, &srv.UserID, &srv.IsActive, &srv.CreatedAt, &srv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := s.decryptRuntime(&srv); err != nil {
		return nil, err
	}
	return &srv, nil
}

func (s *Service) CreateServer(req *CreateServerRequest, userID int) (*Server, error) {
	encUser, encPass, err := s.encryptCredentials(req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	res, err := s.db.Exec(`INSERT INTO servers (name,host,port,type,username_enc,password_enc,user_id,is_active,created_at,updated_at) VALUES (?,?,?,?,?,?,?,1,NOW(),NOW())`, req.Name, req.Host, req.Port, req.Type, encUser, encPass, userID)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetServerByID(int(id), userID)
}

func (s *Service) UpdateServer(id, userID int, req *UpdateServerRequest) (*Server, error) {
	existing, err := s.GetServerByID(id, userID)
	if err != nil {
		return nil, err
	}
	set := []string{}
	args := []interface{}{}
	if req.Name != "" {
		set = append(set, "name=?")
		args = append(args, req.Name)
	}
	if req.Host != "" {
		set = append(set, "host=?")
		args = append(args, req.Host)
	}
	if req.Port != 0 {
		set = append(set, "port=?")
		args = append(args, req.Port)
	}
	if req.Username != "" {
		encUser, _, err := s.encryptCredentials(req.Username, existing.PasswordDecrypted)
		if err != nil {
			return nil, err
		}
		set = append(set, "username_enc=?")
		args = append(args, encUser)
	}
	if req.Password != "" {
		_, encPass, err := s.encryptCredentials(existing.UsernameDecrypted, req.Password)
		if err != nil {
			return nil, err
		}
		set = append(set, "password_enc=?")
		args = append(args, encPass)
	}
	if req.IsActive != nil {
		set = append(set, "is_active=?")
		args = append(args, *req.IsActive)
	}
	if len(set) == 0 {
		return existing, nil
	}
	query := fmt.Sprintf("UPDATE servers SET %s, updated_at=NOW() WHERE id=? AND user_id=?", strings.Join(set, ","))
	args = append(args, id, userID)
	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return s.GetServerByID(id, userID)
}

func (s *Service) DeleteServer(id, userID int) error {
	res, err := s.db.Exec(`UPDATE servers SET is_active=0, updated_at=NOW() WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("server not found")
	}
	return nil
}

// --- Internal helpers ---
func (s *Service) encryptCredentials(username, password string) (string, string, error) {
	key, _ := s.key()
	u, err := crypto.EncryptString(key, username)
	if err != nil {
		return "", "", err
	}
	p, err := crypto.EncryptString(key, password)
	if err != nil {
		return "", "", err
	}
	return u, p, nil
}

func (s *Service) decryptRuntime(srv *Server) error {
	key, _ := s.key()
	u, err := crypto.DecryptString(key, srv.UsernameEnc)
	if err != nil {
		return err
	}
	p, err := crypto.DecryptString(key, srv.PasswordEnc)
	if err != nil {
		return err
	}
	srv.UsernameDecrypted = u
	srv.PasswordDecrypted = p
	return nil
}
