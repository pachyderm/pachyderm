package pps

import(
	"google.golang.org/grpc"
)

func NewInternalJobAPIClientFromAddress(pachAddr string) (InternalJobAPIClient, error) {
	clientConn, err := grpc.Dial(pachAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := NewInternalJobAPIClient(clientConn)

	return client, nil
}
