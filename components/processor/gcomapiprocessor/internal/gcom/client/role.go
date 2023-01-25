package client

// RoleType is used to represent Grafana Cloud roles
type RoleType string

const (
	RoleViewer           RoleType = "Viewer"
	RoleEditor           RoleType = "Editor"
	RoleMetricsPublisher RoleType = "MetricsPublisher"
	RoleAdmin            RoleType = "Admin"
)

// IsValid returns if a RoleType is valid
func (r RoleType) IsValid() bool {
	return r == RoleViewer || r == RoleAdmin || r == RoleEditor || r == RoleMetricsPublisher
}

// IsPublisher determines the Role has publisher privileges
func (r RoleType) IsPublisher() bool {
	return r == RoleMetricsPublisher || r == RoleAdmin || r == RoleEditor
}

// IsViewer determines if the Role has viewer privileges
func (r RoleType) IsViewer() bool {
	return r == RoleViewer || r == RoleAdmin || r == RoleEditor
}

// IsEditor determines if the Role has editor privileges
func (r RoleType) IsEditor() bool {
	return r == RoleAdmin || r == RoleEditor
}

// IsAdmin determines if the Role has admin privileges
func (r RoleType) IsAdmin() bool {
	return r == RoleAdmin
}
