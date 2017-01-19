package version

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"

	"github.com/urfave/cli"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/version"
	"google.golang.org/grpc"
)

// NewCommand returns the version command.
func NewCommand() cli.Command {
	return cmdVersion
}

var cmdVersion = cli.Command{
	Name:        "version",
	Aliases:     []string{"v"},
	Usage:       "Return version information.",
	Description: "Return version information.",
	Action:      actVersion,
}

func actVersion(c *cli.Context) (retErr error) {
	address := c.GlobalString("address")
	if c.GlobalBool("metrics") {
		metricsFn := metrics.ReportAndFlushUserAction("Version")
		defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())
	}
	writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	printVersionHeader(writer)
	printVersion(writer, "pach", version.Version)
	writer.Flush()

	versionClient, err := getVersionAPIClient(address)
	if err != nil {
		return sanitizeErr(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	version, err := versionClient.GetVersion(ctx, &google_protobuf.Empty{})

	if err != nil {
		buf := bytes.NewBufferString("")
		errWriter := tabwriter.NewWriter(buf, 20, 1, 3, ' ', 0)
		fmt.Fprintf(errWriter, "pachd\t(version unknown) : error connecting to pachd server at address (%v): %v\n\n", address, sanitizeErr(err))
		fmt.Fprintf(errWriter, "please make sure pachd is up (`kubectl get all`) and portforwarding is enabled\n")
		errWriter.Flush()
		return errors.New(buf.String())
	}

	printVersion(writer, "pachd", version)
	return writer.Flush()
}

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}

func getVersionAPIClient(address string) (protoversion.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return protoversion.NewAPIClient(clientConn), nil
}

func printVersionHeader(w io.Writer) {
	fmt.Fprintf(w, "COMPONENT\tVERSION\t\n")
}

func printVersion(w io.Writer, component string, v *protoversion.Version) {
	fmt.Fprintf(w, "%s\t%s\t\n", component, version.PrettyPrintVersion(v))
}
