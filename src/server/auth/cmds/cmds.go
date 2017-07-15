package cmds

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra"
)

var githubAuthLink = `http://github.com/login/oauth/authorize?client_id=d3481e92b4f09ea74ff8&redirect_uri=https%3A%2F%2Fpachyderm.io%2Flogin-hook%2Fdisplay-token.html`

// LoginCmd returns a cobra.Command to login to a Pachyderm cluster with your
// GitHub account. Any resources that have been restricted to the email address
// registered with your GitHub account will subsequently be accessible.
func LoginCmd() *cobra.Command {
	login := &cobra.Command{
		Use:   "login",
		Short: "Login to Pachyderm with your GitHub account",
		Long: "Login to Pachyderm with your GitHub account. Any resources that " +
			"have been restricted to the email address registered with your GitHub " +
			"account will subsequently be accessible.",
		Run: cmdutil.Run(func([]string) error {
			// TODO(msteffen): Check if auth service is deployed. If not, print an
			// error and exit
			fmt.Println("(1) Please paste this link into a browser:\n" +
				githubAuthLink + "\n" +
				"(You will be directed to GitHub and asked to authorize Pachyderm's " +
				"login app on Github. If you accept, you will be given a token to " +
				"paste here, which will give you an externally verified account in " +
				"this Pachyderm cluster)\n\n(2) Please paste the token you receive " +
				"from GitHub here:")
			token, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading token: %s", err.Error())
			}
			token = strings.TrimSpace(token) // drop trailing newline
			cfg, err := config.Read()
			if err != nil {
				return fmt.Errorf("error reading Pachyderm config (for cluster "+
					"address): %s", err.Error())
			}
			clientConn, err := grpc.Dial(client.GetAddressFromUserMachine(cfg),
				client.PachDialOptions()...)
			if err != nil {
				return fmt.Errorf("error connecting to Pachyderm cluster: %s",
					err.Error())
			}
			authClient := auth.NewAPIClient(clientConn)

			// Create PachAuth authentication request
			ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
			resp, err := authClient.Authenticate(ctx,
				&auth.AuthenticateRequest{GithubToken: token},
			)
			if err != nil {
				return fmt.Errorf("error authenticating with Pachyderm cluster: %s",
					err.Error())
			}
			cfg.V1.SessionToken = resp.PachToken
			return cfg.Write()
		}),
	}
	return login
}

// Cmds returns a list of cobra commands for authenticating and authorizing
// users in an auth-enabled Pachyderm cluster.
func Cmds() []*cobra.Command {
	return []*cobra.Command{LoginCmd()}
}
