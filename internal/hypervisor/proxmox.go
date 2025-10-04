package hypervisor

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ProxmoxClient struct {
	baseURL   string
	username  string
	password  string
	ticket    string
	csrfToken string
	client    *http.Client
	connected bool
}

type proxmoxAuthResponse struct {
	Data struct {
		Ticket    string `json:"ticket"`
		CSRFToken string `json:"CSRFPreventionToken"`
	} `json:"data"`
}

func NewProxmoxClient() *ProxmoxClient {
	return &ProxmoxClient{client: &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, Timeout: 30 * time.Second}}
}

func (p *ProxmoxClient) GetType() string   { return "prx" }
func (p *ProxmoxClient) IsConnected() bool { return p.connected }
func (p *ProxmoxClient) Disconnect() error {
	p.ticket = ""
	p.csrfToken = ""
	p.connected = false
	return nil
}

func (p *ProxmoxClient) Connect(ctx context.Context, server *Server) error {
	p.baseURL = fmt.Sprintf("https://%s:%d/api2/json", server.Host, server.Port)
	p.username = server.Username
	p.password = server.Password

	data := url.Values{}
	data.Set("username", p.username)
	data.Set("password", p.password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/access/ticket", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed: status %d", resp.StatusCode)
	}

	var ar proxmoxAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return err
	}
	p.ticket = ar.Data.Ticket
	p.csrfToken = ar.Data.CSRFToken
	p.connected = true
	return nil
}

func (p *ProxmoxClient) TestConnection(ctx context.Context) error {
	if !p.connected {
		return ErrConnectionFailed
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/version", nil)
	p.setAuthHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ErrConnectionFailed
	}
	return nil
}

func (p *ProxmoxClient) setAuthHeaders(req *http.Request) {
	if p.ticket != "" {
		req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", p.ticket))
	}
	if p.csrfToken != "" {
		req.Header.Set("CSRFPreventionToken", p.csrfToken)
	}
}

// Получение списка нод
func (p *ProxmoxClient) getNodes(ctx context.Context) ([]string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/nodes", nil)
	p.setAuthHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get nodes status %d", resp.StatusCode)
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	arr, _ := raw["data"].([]any)
	nodes := make([]string, 0, len(arr))
	for _, n := range arr {
		if m, ok := n.(map[string]any); ok {
			if s, ok := m["node"].(string); ok {
				nodes = append(nodes, s)
			}
		}
	}
	return nodes, nil
}

// Универсальная выборка списка инстансов (vm или lxc)
func (p *ProxmoxClient) getInstancesFromNode(ctx context.Context, node, kind string) ([]*Instance, error) {
	endpoint := fmt.Sprintf("/nodes/%s/%s", node, kind)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+endpoint, nil)
	p.setAuthHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s status %d", kind, resp.StatusCode)
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	arr, _ := raw["data"].([]any)
	res := []*Instance{}
	for _, v := range arr {
		m, _ := v.(map[string]any)
		vmid := toInt(m["vmid"]) // для lxc тоже vmid
		name, _ := m["name"].(string)
		status, _ := m["status"].(string)
		nodeName, _ := m["node"].(string)
		cpuPct := int(toFloat(m["cpu"]) * 100)
		mem := bytesToMB(m["mem"], m["maxmem"]) // используем текущую память
		disk := bytesToGB(m["disk"])
		res = append(res, &Instance{ID: strconv.Itoa(vmid), Name: name, Type: map[string]string{"qemu": "vm", "lxc": "lxc"}[kind], Status: status, CPU: cpuPct, RAM: mem, Disk: disk, Node: nodeName})
	}
	return res, nil
}

func (p *ProxmoxClient) GetVMs(ctx context.Context) ([]*Instance, error) {
	nodes, err := p.getNodes(ctx)
	if err != nil {
		return nil, err
	}
	var all []*Instance
	for _, n := range nodes {
		list, err := p.getInstancesFromNode(ctx, n, "qemu")
		if err == nil {
			all = append(all, list...)
		}
	}
	return all, nil
}
func (p *ProxmoxClient) GetLXCs(ctx context.Context) ([]*Instance, error) {
	nodes, err := p.getNodes(ctx)
	if err != nil {
		return nil, err
	}
	var all []*Instance
	for _, n := range nodes {
		list, err := p.getInstancesFromNode(ctx, n, "lxc")
		if err == nil {
			all = append(all, list...)
		}
	}
	return all, nil
}
func (p *ProxmoxClient) GetInstances(ctx context.Context) ([]*Instance, error) {
	vms, err := p.GetVMs(ctx)
	if err != nil {
		return nil, err
	}
	lxcs, err := p.GetLXCs(ctx)
	if err != nil {
		return nil, err
	}
	return append(vms, lxcs...), nil
}

// Действия
func (p *ProxmoxClient) performAction(ctx context.Context, instanceType, instanceID, action string) error {
	node, err := p.findNodeForInstance(ctx, instanceType, instanceID)
	if err != nil {
		return err
	}
	var path string
	if instanceType == "vm" {
		path = fmt.Sprintf("/nodes/%s/qemu/%s/status/%s", node, instanceID, action)
	} else {
		path = fmt.Sprintf("/nodes/%s/lxc/%s/status/%s", node, instanceID, action)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, nil)
	p.setAuthHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ErrActionFailed
	}
	return nil
}

func (p *ProxmoxClient) StartInstance(ctx context.Context, t, id string) error {
	return p.performAction(ctx, t, id, "start")
}
func (p *ProxmoxClient) StopInstance(ctx context.Context, t, id string) error {
	return p.performAction(ctx, t, id, "stop")
}
func (p *ProxmoxClient) RestartInstance(ctx context.Context, t, id string) error {
	return p.performAction(ctx, t, id, "reboot")
}

func (p *ProxmoxClient) GetInstanceStatus(ctx context.Context, t, id string) (string, error) {
	node, err := p.findNodeForInstance(ctx, t, id)
	if err != nil {
		return "", err
	}
	var path string
	if t == "vm" {
		path = fmt.Sprintf("/nodes/%s/qemu/%s/status/current", node, id)
	} else {
		path = fmt.Sprintf("/nodes/%s/lxc/%s/status/current", node, id)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+path, nil)
	p.setAuthHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", ErrInstanceNotFound
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", err
	}
	data, _ := raw["data"].(map[string]any)
	status, _ := data["status"].(string)
	return status, nil
}

func (p *ProxmoxClient) GetInstanceConfig(ctx context.Context, t, id string) (map[string]interface{}, error) {
	node, err := p.findNodeForInstance(ctx, t, id)
	if err != nil {
		return nil, err
	}
	var path string
	if t == "vm" {
		path = fmt.Sprintf("/nodes/%s/qemu/%s/config", node, id)
	} else {
		path = fmt.Sprintf("/nodes/%s/lxc/%s/config", node, id)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+path, nil)
	p.setAuthHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrInstanceNotFound
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	data, _ := raw["data"].(map[string]any)
	return data, nil
}

func (p *ProxmoxClient) DeleteInstance(ctx context.Context, t, id string) error {
	node, err := p.findNodeForInstance(ctx, t, id)
	if err != nil {
		return err
	}
	var path string
	if t == "vm" {
		path = fmt.Sprintf("/nodes/%s/qemu/%s", node, id)
	} else {
		path = fmt.Sprintf("/nodes/%s/lxc/%s", node, id)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, p.baseURL+path, nil)
	p.setAuthHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ErrActionFailed
	}
	return nil
}

func (p *ProxmoxClient) CreateSnapshot(ctx context.Context, t, id, name string) error {
	node, err := p.findNodeForInstance(ctx, t, id)
	if err != nil {
		return err
	}
	var path string
	if t == "vm" {
		path = fmt.Sprintf("/nodes/%s/qemu/%s/snapshot", node, id)
	} else {
		path = fmt.Sprintf("/nodes/%s/lxc/%s/snapshot", node, id)
	}
	data := url.Values{}
	data.Set("snapname", name)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	p.setAuthHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ErrActionFailed
	}
	return nil
}

// Поиск ноды по инстансу
func (p *ProxmoxClient) findNodeForInstance(ctx context.Context, t, id string) (string, error) {
	nodes, err := p.getNodes(ctx)
	if err != nil {
		return "", err
	}
	for _, n := range nodes {
		var path string
		if t == "vm" {
			path = fmt.Sprintf("/nodes/%s/qemu/%s", n, id)
		} else {
			path = fmt.Sprintf("/nodes/%s/lxc/%s", n, id)
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+path, nil)
		p.setAuthHeaders(req)
		resp, err := p.client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return n, nil
		}
	}
	return "", ErrInstanceNotFound
}

// Helpers
func toInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case json.Number:
		i, _ := x.Int64()
		return int(i)
	default:
		return 0
	}
}
func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case json.Number:
		f, _ := x.Float64()
		return f
	default:
		return 0
	}
}
func bytesToMB(cur any, _ any) int { // упрощено
	return int(toFloat(cur) / 1024.0 / 1024.0)
}
func bytesToGB(cur any) int { return int(toFloat(cur) / 1024.0 / 1024.0 / 1024.0) }
