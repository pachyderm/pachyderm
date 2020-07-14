package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type PermissionTestable interface {
	Run() error
}

type permissionTest struct {
	input *Permission
	i     int
}

// Describes test case for Permission.Set
type SetPermissionTest struct {
	permissionTest
	expected uint64
}

// Testable interface for Permission.Set
func (spt *SetPermissionTest) Run() error {
	spt.input.Set(spt.i)
	if spt.input.Perm != spt.expected {
		return fmt.Errorf("Expected %x, Got %x", spt.expected, spt.input.Perm)
	}
	return nil
}

// Describes test case for Permission.Clear
type ClearPermissionTest struct {
	permissionTest
	expected uint64
}

// Testable interface for Permission.Clear
func (cpt *ClearPermissionTest) Run() error {
	cpt.input.Clear(cpt.i)
	if cpt.input.Perm != cpt.expected {
		return fmt.Errorf("Expected %x, Got %x", cpt.expected, cpt.input.Perm)
	}
	return nil
}

// Describes test case for Permission.Test
type TestPermissionTest struct {
	permissionTest
	expected bool
}

// Testable interface for Permission.Test
func (tpt *TestPermissionTest) Run() error {
	res := tpt.input.Test(tpt.i)
	if res != tpt.expected {
		return fmt.Errorf("Expected %t, Got %t", tpt.expected, res)
	}
	return nil
}

var permissionTests = []PermissionTestable{
	// lsb
	&SetPermissionTest{
		permissionTest{
			input: &Permission{Perm: uint64(0x0)},
			i:     0,
		},
		uint64(0x1),
	},
	// msb
	&SetPermissionTest{
		permissionTest{
			input: &Permission{Perm: uint64(0x0)},
			i:     63,
		},
		uint64(0x8000000000000000),
	},
	&ClearPermissionTest{
		permissionTest{
			input: &Permission{Perm: uint64(0x8000000000000000)},
			i:     63,
		},
		uint64(0x0),
	},
	&ClearPermissionTest{
		permissionTest{
			input: &Permission{Perm: uint64(0x8)},
			i:     3,
		},
		uint64(0x0),
	},
	&TestPermissionTest{
		permissionTest{
			input: &Permission{Perm: uint64(0xf)},
			i:     3,
		},
		true,
	},
	&TestPermissionTest{
		permissionTest{
			input: &Permission{Perm: uint64(0x0)},
			i:     4,
		},
		false,
	},
	&TestPermissionTest{
		permissionTest{
			input: &Permission{Perm: uint64(0xfe)},
			i:     0,
		},
		false,
	},
}

// Run all tests in above list
func TestRunPermissionTests(t *testing.T) {
	// register permissions types
	PermissionsRegistry.Register("testuser", NewPermissionsBitmap("admin", "edit", "billing"))

	for i, z := range permissionTests {
		if err := z.Run(); err != nil {
			t.Error(fmt.Sprintf("Test %d", i), err)
		}
	}
}

func TestPermissionsHighBits(t *testing.T) {
	// register permissions types
	PermissionsRegistry.Register("testuser", NewPermissionsBitmap("admin", "edit", "billing"))

	p := &Permission{Perm: uint64(0x3), Name: "testuser"}

	expected := []int{0, 1}
	res := p.HighBits()
	if !reflect.DeepEqual(expected, res) {
		t.Error(fmt.Sprintf("TestPermissionsHighBits expected %v, got %v", expected, res))
	}
}

type SetPermissionsTest struct {
	perms    *Permission // 0x0
	pnames   []string    // to set
	expected []string    // .Permissions()
	failed   []string    // invalid permissions
	op       string      // set or clear
}

var setPermissionsTests = []*SetPermissionsTest{
	&SetPermissionsTest{
		perms:    &Permission{Perm: uint64(0x0), Name: "testuser"},
		pnames:   []string{"admin", "edit", "billing"},
		expected: []string{"admin", "edit", "billing"},
		op:       "set",
	},
	&SetPermissionsTest{
		perms:    &Permission{Perm: uint64(0x7), Name: "testuser"},
		pnames:   []string{"admin"},
		expected: []string{"edit", "billing"},
		op:       "clear",
	},
	&SetPermissionsTest{
		perms:    &Permission{Perm: uint64(0x3), Name: "testuser"},
		pnames:   []string{"billing"},
		expected: []string{"admin", "edit", "billing"},
		op:       "set",
	},
	&SetPermissionsTest{
		perms:    &Permission{Perm: uint64(0x3), Name: "testuser"},
		pnames:   []string{"invalid"},
		expected: []string{"admin", "edit"},
		failed:   []string{"invalid"},
		op:       "set",
	},
}

// Tests set and clear of permissions
func TestSetPermissions(t *testing.T) {
	// register permissions types
	PermissionsRegistry.Register("testuser", NewPermissionsBitmap("admin", "edit", "billing"))

	for i, test := range setPermissionsTests {
		var failed []string
		switch test.op {
		case "set":
			failed = test.perms.SetPermissions(test.pnames...)
		case "clear":
			failed = test.perms.ClearPermissions(test.pnames...)
		}
		res := test.perms.Permissions()
		if !reflect.DeepEqual(test.expected, res) {
			t.Error(fmt.Sprintf("TestSetPermission #%d expected valid perms %v, got %v", i, test.expected, res))
		}
		if !reflect.DeepEqual(test.failed, failed) {
			t.Error(fmt.Sprintf("TestSetPermission #%d expected invalid perms %v, got %v", i, test.failed, failed))
		}

	}
}

func TestPermissionsMarshalJSON(t *testing.T) {
	// register permissions types
	PermissionsRegistry.Register("testuser", NewPermissionsBitmap("admin", "edit", "billing"))

	// test marshalling of json 011
	expected := []string{"admin", "edit"}
	jb, _ := json.Marshal(&Permission{Perm: uint64(0x3), Name: "testuser"})

	var res []string
	_ = json.Unmarshal(jb, &res)
	if !reflect.DeepEqual(expected, res) {
		t.Error(fmt.Sprintf("TestPermissionsMarshalJSON expected %v, got %v", expected, res))
	}
}

type CheckPermissionsTest struct {
	perms    *Permission
	pnames   []string
	expected map[string]*Error
}

var permissionsTests = []*CheckPermissionsTest{
	&CheckPermissionsTest{
		perms:  &Permission{Perm: uint64(0x7), Name: "testuser"},
		pnames: []string{"admin", "edit", "billing"},
		expected: map[string]*Error{
			"admin":   nil,
			"edit":    nil,
			"billing": nil,
		},
	},
	&CheckPermissionsTest{
		perms:  &Permission{Perm: uint64(0x0), Name: "testuser"},
		pnames: []string{"admin", "edit", "billing"},
		expected: map[string]*Error{
			"admin":   NewPermissionsError("admin"),
			"edit":    NewPermissionsError("edit"),
			"billing": NewPermissionsError("billing"),
		},
	},
	&CheckPermissionsTest{
		perms:  &Permission{Perm: uint64(0x2), Name: "testuser"},
		pnames: []string{"admin", "edit", "billing"},
		expected: map[string]*Error{
			"admin":   NewPermissionsError("admin"),
			"edit":    nil,
			"billing": NewPermissionsError("billing"),
		},
	},
}

func TestCheckPermissionsTests(t *testing.T) {
	// register permissions types
	PermissionsRegistry.Register("testuser", NewPermissionsBitmap("admin", "edit", "billing"))

	for i, test := range permissionsTests {
		output := test.perms.CheckPermissions(test.pnames...)
		if len(output) != len(test.expected) {
			t.Error(fmt.Sprintf("Test %d, expected len %d, got len %d", i, len(test.expected), len(output)))
			continue
		}
		for k, errexpected := range test.expected {
			if errgot, ok := output[k]; ok {
				if errgot == nil && errexpected != nil {
					t.Error(fmt.Sprintf("Test %d, expected \"%v\", got nil", i, errexpected.Error()))
				} else if errgot != nil && errexpected == nil {
					t.Error(fmt.Sprintf("Test %d, expected nil, got \"%s\"", i, errgot.Error()))
				} else if errgot != nil && errexpected != nil {
					if errexpected.Error() != errgot.Error() {
						t.Error(fmt.Sprintf("Test %d, expected \"%s\", got %s\n\n", i, errexpected.Error(), errgot.Error()))
					}
				}
			} else {
				t.Error(fmt.Sprintf("Test %d, expected %s, got \"key did not exist\"", i, errexpected))
			}
		}
	}
}
