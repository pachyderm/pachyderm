package s3server

// User is an XML-encodable representation of an S3 user
type User struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}
