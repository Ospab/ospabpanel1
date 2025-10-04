package vps

import (
	"database/sql"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetVPSByUserID(userID int) ([]*VPS, error) {
	query := "SELECT id, name, status, ip_address, cpu, ram, disk, os, user_id, hypervisor, created_at, updated_at FROM vps WHERE user_id = ?"

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vpsList []*VPS
	for rows.Next() {
		vps := &VPS{}
		err := rows.Scan(
			&vps.ID,
			&vps.Name,
			&vps.Status,
			&vps.IPAddress,
			&vps.CPU,
			&vps.RAM,
			&vps.Disk,
			&vps.OS,
			&vps.UserID,
			&vps.Hypervisor,
			&vps.CreatedAt,
			&vps.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		vpsList = append(vpsList, vps)
	}

	return vpsList, nil
}

func (s *Service) GetVPSByID(id int) (*VPS, error) {
	vps := &VPS{}
	query := "SELECT id, name, status, ip_address, cpu, ram, disk, os, user_id, hypervisor, created_at, updated_at FROM vps WHERE id = ?"

	err := s.db.QueryRow(query, id).Scan(
		&vps.ID,
		&vps.Name,
		&vps.Status,
		&vps.IPAddress,
		&vps.CPU,
		&vps.RAM,
		&vps.Disk,
		&vps.OS,
		&vps.UserID,
		&vps.Hypervisor,
		&vps.CreatedAt,
		&vps.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return vps, nil
}

func (s *Service) CreateVPS(req *CreateVPSRequest, userID int) (*VPS, error) {
	query := `INSERT INTO vps (name, status, cpu, ram, disk, os, user_id, hypervisor, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`

	result, err := s.db.Exec(query, req.Name, "creating", req.CPU, req.RAM, req.Disk, req.OS, userID, req.Hypervisor)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetVPSByID(int(id))
}

func (s *Service) UpdateVPSStatus(id int, status string) error {
	query := "UPDATE vps SET status = ?, updated_at = NOW() WHERE id = ?"
	_, err := s.db.Exec(query, status, id)
	return err
}

func (s *Service) DeleteVPS(id int) error {
	query := "DELETE FROM vps WHERE id = ?"
	_, err := s.db.Exec(query, id)
	return err
}
