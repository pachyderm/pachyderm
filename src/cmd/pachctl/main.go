package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm"
	pfscmds "github.com/pachyderm/pachyderm/src/pfs/cmds"
	deploycmds "github.com/pachyderm/pachyderm/src/pkg/deploy/cmds"
	ppscmds "github.com/pachyderm/pachyderm/src/pps/cmds"
	"github.com/spf13/cobra"
	"go.pedge.io/env"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/proto/version"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS,default=0.0.0.0:650"`
	PachydermPpsd1Port string `env:"PACHYDERM_PPSD_1_PORT"`
	PpsAddress         string `env:"PPS_ADDRESS,default=0.0.0.0:651"`
	KubernetesAddress  string `env:"KUBERNETES_ADDRESS,default=http://localhost:8080"`
	KubernetesUsername string `env:"KUBERNETES_USERNAME,default=admin"`
	KubernetesPassword string `env:"KUBERNETES_PASSWORD"`
	Provider           string `env:"PROVIDER"`
	GCEProject         string `env:"GCE_PROJECT"`
	GCEZone            string `env:"GCE_ZONE"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Envronment variables:
  PFS_ADDRESS=0.0.0.0:650, the PFS server to connect to.
  PPS_ADDRESS=0.0.0.0:651, the PPS server to connect to.
  KUBERNETES_ADDRESS=http://localhost:8080, the Kubernetes endpoint to connect to.
  KUBERNETES_USERNAME=admin
  KUBERNETES_PASSWORD
  PROVIDER, which provider to use for cluster creation (currently only supports GCE).
  GCE_PROJECT
  GCE_ZONE`,
	}
	pfsdAddress := getPfsdAddress(appEnv)
	ppsdAddress := getPpsdAddress(appEnv)
	pfsCmds, err := pfscmds.Cmds(pfsdAddress)
	if err != nil {
		return err
	}
	for _, cmd := range pfsCmds {
		rootCmd.AddCommand(cmd)
	}
	ppsCmds, err := ppscmds.Cmds(ppsdAddress)
	if err != nil {
		return err
	}
	for _, cmd := range ppsCmds {
		rootCmd.AddCommand(cmd)
	}

	deployCmds, err := deploycmds.Cmds(
		appEnv.KubernetesAddress,
		appEnv.KubernetesUsername,
		appEnv.KubernetesAddress,
		appEnv.Provider,
		appEnv.GCEProject,
		appEnv.GCEZone,
	)
	if err != nil {
		return err
	}
	for _, cmd := range deployCmds {
		rootCmd.AddCommand(cmd)
	}
	version := &cobra.Command{
		Use:   "version",
		Short: "Return version information.",
		Long:  "Return version information.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			pfsdVersionClient, err := getVersionAPIClient(pfsdAddress)
			if err != nil {
				return err
			}
			pfsdVersion, err := pfsdVersionClient.GetVersion(context.Background(), &google_protobuf.Empty{})
			if err != nil {
				return err
			}
			ppsdVersionClient, err := getVersionAPIClient(pfsdAddress)
			if err != nil {
				return err
			}
			ppsdVersion, err := ppsdVersionClient.GetVersion(context.Background(), &google_protobuf.Empty{})
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			printVerisonHeader(writer)
			printVersion(writer, "pachctl", pachyderm.Version)
			printVersion(writer, "pfsd", pfsdVersion)
			printVersion(writer, "ppsd", ppsdVersion)
			return writer.Flush()
		}),
	}
	rootCmd.AddCommand(version)

	return rootCmd.Execute()
}

func getPfsdAddress(appEnv *appEnv) string {
	if pfsdAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR"); pfsdAddr != "" {
		return fmt.Sprintf("%s:650", pfsdAddr)
	}
	if appEnv.PachydermPfsd1Port != "" {
		return strings.Replace(appEnv.PachydermPfsd1Port, "tcp://", "", -1)
	}
	return appEnv.PfsAddress
}

func getPpsdAddress(appEnv *appEnv) string {
	if ppsdAddr := os.Getenv("PPSD_PORT_651_TCP_ADDR"); ppsdAddr != "" {
		return fmt.Sprintf("%s:651", ppsdAddr)
	}
	if appEnv.PachydermPpsd1Port != "" {
		return strings.Replace(appEnv.PachydermPpsd1Port, "tcp://", "", -1)
	}
	return appEnv.PpsAddress
}

func getVersionAPIClient(address string) (protoversion.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return protoversion.NewAPIClient(clientConn), nil
}

func printVerisonHeader(w io.Writer) {
	fmt.Fprintf(w, "COMPONENT\tVERSION\t\n")
}

func printVersion(w io.Writer, component string, version *protoversion.Version) {
	fmt.Fprintf(
		w,
		"%s\t%d.%d.%d(%s)\t\n",
		component,
		version.Major,
		version.Minor,
		version.Micro,
		version.Additional,
	)
}
