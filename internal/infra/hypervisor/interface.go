package hypervisor

import (
	"ospab-panel/internal/core/vps"
)

type Hypervisor interface {
	CreateVPS(req *vps.CreateVPSRequest) (*vps.VPS, error)
	StartVPS(vpsID string) error
	StopVPS(vpsID string) error
	RestartVPS(vpsID string) error
	DeleteVPS(vpsID string) error
	GetVPSStatus(vpsID string) (string, error)
	GetVPSInfo(vpsID string) (*vps.VPS, error)
}

type Config struct {
	Type     string `json:"type"` // kvm, docker, etc
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}
