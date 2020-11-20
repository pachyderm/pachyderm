package server

import (
	"fmt"
	"time"

	"github.com/dexidp/dex/storage"
)

// proxyStorage extends dex's built-in storage with Pachyderm-specific features
type proxyStorage struct {
	storage storage.Storage
}

// ListConnectors returns a fake connector if none are configured, so the server
// starts successfully
func (s *proxyStorage) ListConnectors() ([]storage.Connector, error) {
	c, err := s.storage.ListConnectors()
	if err != nil {
		return nil, err
	}
	fmt.Printf("got connectors: %v\n", c)

	if len(c) == 0 {
		fmt.Printf("injecting fake\n")
		// This connector isn't actually available via `GetConnector` so it can't be used to authenticate
		return []storage.Connector{storage.Connector{
			ID:              "placeholder",
			Type:            "mockPassword",
			Name:            "No storage.Connectors Configured",
			ResourceVersion: "",
			Config:          []byte(`{"username": "unused", "password": "unused"}`),
		}}, nil
	}
	return c, err
}
func (s *proxyStorage) Close() error {
	return s.storage.Close()
}
func (s *proxyStorage) CreateAuthRequest(a storage.AuthRequest) error {
	return s.storage.CreateAuthRequest(a)
}
func (s *proxyStorage) CreateClient(c storage.Client) error {
	return s.storage.CreateClient(c)
}
func (s *proxyStorage) CreateAuthCode(c storage.AuthCode) error {
	return s.storage.CreateAuthCode(c)
}
func (s *proxyStorage) CreateRefresh(r storage.RefreshToken) error {
	return s.storage.CreateRefresh(r)
}
func (s *proxyStorage) CreatePassword(p storage.Password) error {
	return s.storage.CreatePassword(p)
}
func (s *proxyStorage) CreateOfflineSessions(ss storage.OfflineSessions) error {
	return s.storage.CreateOfflineSessions(ss)
}
func (s *proxyStorage) CreateConnector(c storage.Connector) error {
	return s.storage.CreateConnector(c)
}
func (s *proxyStorage) CreateDeviceRequest(d storage.DeviceRequest) error {
	return s.storage.CreateDeviceRequest(d)
}
func (s *proxyStorage) CreateDeviceToken(d storage.DeviceToken) error {
	return s.storage.CreateDeviceToken(d)
}
func (s *proxyStorage) GetAuthRequest(id string) (storage.AuthRequest, error) {
	return s.storage.GetAuthRequest(id)
}
func (s *proxyStorage) GetAuthCode(id string) (storage.AuthCode, error) {
	return s.storage.GetAuthCode(id)
}
func (s *proxyStorage) GetClient(id string) (storage.Client, error) {
	return s.storage.GetClient(id)
}
func (s *proxyStorage) GetKeys() (storage.Keys, error) {
	return s.storage.GetKeys()
}
func (s *proxyStorage) GetRefresh(id string) (storage.RefreshToken, error) {
	return s.storage.GetRefresh(id)
}
func (s *proxyStorage) GetPassword(email string) (storage.Password, error) {
	return s.storage.GetPassword(email)
}
func (s *proxyStorage) GetOfflineSessions(userID string, connID string) (storage.OfflineSessions, error) {
	return s.storage.GetOfflineSessions(userID, connID)
}
func (s *proxyStorage) GetConnector(id string) (storage.Connector, error) {
	return s.storage.GetConnector(id)
}
func (s *proxyStorage) GetDeviceRequest(userCode string) (storage.DeviceRequest, error) {
	return s.storage.GetDeviceRequest(userCode)
}
func (s *proxyStorage) GetDeviceToken(deviceCode string) (storage.DeviceToken, error) {
	return s.storage.GetDeviceToken(deviceCode)
}
func (s *proxyStorage) ListClients() ([]storage.Client, error) {
	return s.storage.ListClients()
}

func (s *proxyStorage) ListRefreshTokens() ([]storage.RefreshToken, error) {
	return s.storage.ListRefreshTokens()
}
func (s *proxyStorage) ListPasswords() ([]storage.Password, error) {
	return s.storage.ListPasswords()
}
func (s *proxyStorage) DeleteAuthRequest(id string) error {
	return s.storage.DeleteAuthRequest(id)
}
func (s *proxyStorage) DeleteAuthCode(code string) error {
	return s.storage.DeleteAuthCode(code)
}
func (s *proxyStorage) DeleteClient(id string) error {
	return s.storage.DeleteClient(id)
}
func (s *proxyStorage) DeleteRefresh(id string) error {
	return s.storage.DeleteRefresh(id)
}
func (s *proxyStorage) DeletePassword(email string) error {
	return s.storage.DeletePassword(email)
}
func (s *proxyStorage) DeleteOfflineSessions(userID string, connID string) error {
	return s.storage.DeleteOfflineSessions(userID, connID)
}
func (s *proxyStorage) DeleteConnector(id string) error {
	return s.storage.DeleteConnector(id)
}
func (s *proxyStorage) UpdateClient(id string, updater func(old storage.Client) (storage.Client, error)) error {
	return s.storage.UpdateClient(id, updater)
}
func (s *proxyStorage) UpdateKeys(updater func(old storage.Keys) (storage.Keys, error)) error {
	return s.storage.UpdateKeys(updater)
}
func (s *proxyStorage) UpdateAuthRequest(id string, updater func(a storage.AuthRequest) (storage.AuthRequest, error)) error {
	return s.storage.UpdateAuthRequest(id, updater)
}
func (s *proxyStorage) UpdateRefreshToken(id string, updater func(r storage.RefreshToken) (storage.RefreshToken, error)) error {
	return s.storage.UpdateRefreshToken(id, updater)
}
func (s *proxyStorage) UpdatePassword(email string, updater func(p storage.Password) (storage.Password, error)) error {
	return s.storage.UpdatePassword(email, updater)
}
func (s *proxyStorage) UpdateOfflineSessions(userID string, connID string, updater func(s storage.OfflineSessions) (storage.OfflineSessions, error)) error {
	return s.storage.UpdateOfflineSessions(userID, connID, updater)
}
func (s *proxyStorage) UpdateConnector(id string, updater func(c storage.Connector) (storage.Connector, error)) error {
	return s.storage.UpdateConnector(id, updater)
}
func (s *proxyStorage) UpdateDeviceToken(deviceCode string, updater func(t storage.DeviceToken) (storage.DeviceToken, error)) error {
	return s.storage.UpdateDeviceToken(deviceCode, updater)
}
func (s *proxyStorage) GarbageCollect(now time.Time) (storage.GCResult, error) {
	return s.storage.GarbageCollect(now)
}
