package misc

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
	"go.pedge.io/pkg/exec"
)

func newPortForwardCommand() cli.Command {

	return cli.Command{
		Name:        "port-forwrad",
		Usage:       "Forward a port on the local machine to pachd. This command blocks.",
		Description: "Forward a port on the local machine to pachd. This command blocks.",
		Action:      actPortForward,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "port, p",
				Usage: "The local port to bind to.",
				Value: 30650,
			},
		},
	}
}

func actPortForward(c *cli.Context) error {
	stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl get pod -l app=pachd | awk '{if (NR!=1) { print $1; exit 0 }}')
kubectl port-forward "$pod" %d:650
`, c.Int("port")))
	fmt.Println("Port forwarded, CTRL-C to exit.")
	return pkgexec.RunIO(pkgexec.IO{Stdin: stdin, Stderr: os.Stderr}, "sh")
}
