package protoclient

import (
	"fmt"
	"os"

	"golang.org/x/net/context"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/version"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NewVersionCommand creates a new command to print the version of the client and server.
func NewVersionCommand(
	clientConn *grpc.ClientConn,
	clientVersion *protoversion.Version,
	errorHandler func(error),
) *cobra.Command {
	return &cobra.Command{
		Use:  "version",
		Long: "Print the version.",
		Run: func(cmd *cobra.Command, args []string) {
			serverVersion, err := protoversion.NewAPIClient(clientConn).GetVersion(
				context.Background(),
				&google_protobuf.Empty{},
			)
			if err != nil {
				if errorHandler != nil {
					errorHandler(err)
					return
				}
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}
			fmt.Printf("Client: %s\nServer: %s\n", clientVersion.VersionString(), serverVersion.VersionString())
		},
	}
}
