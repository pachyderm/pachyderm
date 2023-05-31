package pfs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProject_ValidateName(t *testing.T) {
	var p = &Project{Name: "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"}
	err := p.ValidateName()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf("is %d characters", len(p.Name)-projectNameLimit)), fmt.Sprintf("missing %d", len(p.Name)-projectNameLimit))
}
