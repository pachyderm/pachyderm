package s2

// User is an XML marshallable representation of an S3 user
type User struct {
	// ID is an ID of the user
	ID string `xml:"ID"`
	// DisplayName is a display name of the user
	DisplayName string `xml:"DisplayName"`
}
