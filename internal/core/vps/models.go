package vps

import (
	"time"
)

type VPS struct {
	ID         int       `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Status     string    `json:"status" db:"status"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	CPU        int       `json:"cpu" db:"cpu"`
	RAM        int       `json:"ram" db:"ram"`   // в MB
	Disk       int       `json:"disk" db:"disk"` // в GB
	OS         string    `json:"os" db:"os"`
	UserID     int       `json:"user_id" db:"user_id"`
	Hypervisor string    `json:"hypervisor" db:"hypervisor"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type CreateVPSRequest struct {
	Name       string `json:"name"`
	CPU        int    `json:"cpu"`
	RAM        int    `json:"ram"`
	Disk       int    `json:"disk"`
	OS         string `json:"os"`
	Hypervisor string `json:"hypervisor"`
}

type VPSAction struct {
	Action string `json:"action"` // start, stop, restart, delete
}
