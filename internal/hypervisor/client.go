package hypervisor

import (
	"context"
	"errors"
	"fmt"
)

// Instance представляет VM или LXC контейнер
type Instance struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"` // vm или lxc
	Status string `json:"status"`
	CPU    int    `json:"cpu"`
	RAM    int    `json:"ram"`
	Disk   int    `json:"disk"`
	OS     string `json:"os"`
	Node   string `json:"node"`
}

// Server данные для подключения гипервизора
type Server struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Type     string `json:"type"` // prx, vmv, hyv, kvm, xen
	Username string `json:"username"`
	Password string `json:"-"`
	UserID   int    `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// HypervisorClient интерфейс работы с гипервизорами
// Для расширяемости добавлены общие методы
// Можно будет реализовать для других типов
type HypervisorClient interface {
	Connect(ctx context.Context, server *Server) error
	Disconnect() error
	TestConnection(ctx context.Context) error

	GetInstances(ctx context.Context) ([]*Instance, error)
	GetVMs(ctx context.Context) ([]*Instance, error)
	GetLXCs(ctx context.Context) ([]*Instance, error)

	StartInstance(ctx context.Context, instanceType, instanceID string) error
	StopInstance(ctx context.Context, instanceType, instanceID string) error
	RestartInstance(ctx context.Context, instanceType, instanceID string) error
	GetInstanceStatus(ctx context.Context, instanceType, instanceID string) (string, error)
	GetInstanceConfig(ctx context.Context, instanceType, instanceID string) (map[string]interface{}, error)
	DeleteInstance(ctx context.Context, instanceType, instanceID string) error
	CreateSnapshot(ctx context.Context, instanceType, instanceID, name string) error

	GetType() string
	IsConnected() bool
}

// Фабрика гипервизоров
// Позволит выбирать реализацию по типу

type HypervisorFactory struct{}

func NewHypervisorFactory() *HypervisorFactory { return &HypervisorFactory{} }

func (f *HypervisorFactory) CreateClient(t string) (HypervisorClient, error) {
	switch t {
	case "prx":
		return NewProxmoxClient(), nil
	// Заглушки для будущих реализаций
	case "vmv":
		return nil, errors.New("vmware client not implemented yet")
	case "hyv":
		return nil, errors.New("hyper-v client not implemented yet")
	case "kvm":
		return nil, errors.New("kvm client not implemented yet")
	case "xen":
		return nil, errors.New("xen client not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported hypervisor type: %s", t)
	}
}

var (
	ErrUnsupportedHypervisor = errors.New("unsupported hypervisor type")
	ErrConnectionFailed      = errors.New("connection failed")
	ErrAuthenticationFailed  = errors.New("authentication failed")
	ErrInstanceNotFound      = errors.New("instance not found")
	ErrActionFailed          = errors.New("action failed")
)
