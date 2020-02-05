package config

// EqualClusterReference returns whether two contexts appear to point to the
// same underlying kubernetes cluster
func (c *Context) EqualClusterReference(other *Context) bool {
	if other == nil {
		return false
	}
	if c.Source != other.Source {
		return false
	}
	if c.ClusterName != other.ClusterName {
		return false
	}
	if c.AuthInfo != other.AuthInfo {
		return false
	}
	if c.Namespace != other.Namespace {
		return false
	}
	return true
}
