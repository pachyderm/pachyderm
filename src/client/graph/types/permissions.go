package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"
)

// the permission which corresponds to opsee administrator
const OpseeAdmin = "opsee_admin"

var PermissionsRegistry = NewPermsRegistry()

type PermissionsBitmap struct {
	Name     string
	bitmap   map[int]string
	inverted map[string]int
	sync.RWMutex
}

// Returns new permissions bitmap with mapping of bit to perms in order
func NewPermissionsBitmap(perms ...string) *PermissionsBitmap {
	m := &PermissionsBitmap{
		bitmap:   make(map[int]string),
		inverted: make(map[string]int),
	}
	for i, perm := range perms {
		m.Register(i, perm)
	}
	return m
}

// Returns bit number of a permission
func (p *PermissionsBitmap) Bit(perm string) (int, bool) {
	p.RLock()
	defer p.RUnlock()
	i, ok := p.inverted[perm]
	return i, ok
}

// Returns flag given bitnumber
func (p *PermissionsBitmap) Get(i int) (string, bool) {
	p.RLock()
	defer p.RUnlock()
	t, ok := p.bitmap[i]
	return t, ok
}

//  Returns number of permissions in the Bitmap
func (p *PermissionsBitmap) Length() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.bitmap)
}

// Registers permission in the Bitmap and the Inverted map
func (p *PermissionsBitmap) Register(i int, name string) {
	p.Lock()
	defer p.Unlock()
	p.bitmap[i] = name
	p.inverted[name] = i
}

// Returns new permissions with mapping of bit to perms in order
func NewPermissions(name string, perms ...string) (*Permission, error) {
	p := &Permission{
		Name: name,
	}
	failed := p.SetPermissions(perms...)
	if len(failed) > 0 {
		return nil, NewInvalidPermissionsError(failed...)
	}
	return p, nil
}

// Set flag i in permission
func (p *Permission) Set(i int) {
	p.Perm |= (uint64(1) << uint(i))
}

// UnSet flag i in permission
func (p *Permission) Clear(i int) {
	p.Perm &= ^(uint64(1) << uint(i))
}

// Flag i in permission contains 1
func (p *Permission) Test(i int) bool {
	return (p.Perm&(uint64(1)<<uint(i)) > 0)
}

// Returns dank bits
func (p *Permission) HighBits() []int {
	var hb []int
	for i := 0; i < 64; i++ {
		if p.Test(i) {
			hb = append(hb, i)
		}
	}
	return hb
}

// Returns a list of permissions we have
func (p *Permission) Permissions() []string {
	var perms []string
	r, ok := PermissionsRegistry.Get(p.Name)
	if !ok {
		return []string{}
	}
	for _, bit := range p.HighBits() {
		if perm, ok := r.Get(bit); ok {
			perms = append(perms, perm)
		}
	}
	return perms
}

// Set permission by name
func (p *Permission) SetPermission(perm string) bool {
	reg, rok := PermissionsRegistry.Get(p.Name)
	if !rok {
		return false
	}
	if i, ok := reg.Bit(perm); ok {
		p.Set(i)
		return true
	}
	return false
}

// Set permissions by name
func (p *Permission) SetPermissions(perms ...string) (failed []string) {
	for _, perm := range perms {
		if p.SetPermission(perm) == false {
			failed = append(failed, perm)
		}
	}
	return failed
}

// Clear permission by name
func (p *Permission) ClearPermission(perm string) bool {
	reg, rok := PermissionsRegistry.Get(p.Name)
	if !rok {
		return false
	}
	if i, ok := reg.Bit(perm); ok {
		p.Clear(i)
		return true
	}
	return false
}

// Clear permissions by name
func (p *Permission) ClearPermissions(perms ...string) (failed []string) {
	for _, perm := range perms {
		if p.ClearPermission(perm) == false {
			failed = append(failed, perm)
		}
	}
	return failed
}

// Checks permissions map for permission names, returns errors for those that do not exist
func (p *Permission) HasPermissions(pnames ...string) map[string]bool {
	hasPerms := make(map[string]bool)
	retPerms := make(map[string]bool)
	for _, p := range p.Permissions() {
		hasPerms[p] = true
	}

	for _, name := range pnames {
		if _, ok := hasPerms[name]; ok {
			retPerms[name] = true
		} else {
			retPerms[name] = false
		}
	}
	return retPerms
}

// Checks permissions map for permission names, returns errors for those that do not exist
func (p *Permission) CheckPermissions(pnames ...string) map[string]error {
	permErrs := make(map[string]error)
	for name, has := range p.HasPermissions(pnames...) {
		if !has {
			permErrs[name] = NewPermissionsError(name)
		} else {
			permErrs[name] = nil
		}
	}
	return permErrs
}

// Override MarshalJson to return a list of permissions
func (p *Permission) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Permissions())
}

func (p *Permission) Scan(src interface{}) error {
	switch value := src.(type) {
	case int:
		p.Perm = uint64(value)
	case int32:
		p.Perm = uint64(value)
	case int64:
		p.Perm = uint64(value)
	default:
		return fmt.Errorf("invalid type")
	}

	return p.Validate()
}

func (p *Permission) Validate() error {
	return nil
}

func (p *Permission) Value() (driver.Value, error) {
	return int64(p.Perm), nil
}

func NewPermissionsError(pnames ...string) *Error {
	return NewError("PermissionsError", fmt.Sprintf("Not authorized: %v", pnames))
}

func NewInvalidPermissionsError(pnames ...string) *Error {
	return NewError("PermissionsError", fmt.Sprintf("Invalid: %v", pnames))
}
