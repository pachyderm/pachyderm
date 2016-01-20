package obj

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"golang.org/x/net/context"
)

type amazonClient struct {
	ctx context.Context
}

func newAmazonClient(bucket string, id string, secret string, token string) (*amazonClient, error) {
	credentials := credentials.NewStaticCredentials(id, secret, token)
	session.New(&aws.Config{
		Credentials: credentials,
	})
	return nil, nil
}

func (c *amazonClient) Writer(name string) (io.WriteCloser, error) {
	return nil, nil
}

func (c *amazonClient) Walk(name string, fn func(name string) error) error {
	return fmt.Errorf("amazonClient.Walk: not implemented")
}

//TODO size 0 means read all
func (c *amazonClient) Reader(name string, offset uint64, size uint64) (io.ReadCloser, error) {
	return nil, nil
}

func (c *amazonClient) Delete(name string) error {
	return nil
}
