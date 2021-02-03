package cmds

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pkg/browser"

	"github.com/spf13/cobra"
)

func requestOIDCLogin(c *client.APIClient, openBrowser bool) (string, error) {
	var authURL string
	loginInfo, err := c.GetOIDCLogin(c.Ctx(), &auth.GetOIDCLoginRequest{})
	if err != nil {
		return "", err
	}
	authURL = loginInfo.LoginURL
	state := loginInfo.State

	// print the prepared URL and promp the user to click on it
	fmt.Println("You will momentarily be directed to your IdP and asked to authorize Pachyderm's " +
		"login app on your IdP.\n\nPaste the following URL into a browser if not automatically redirected:\n\n" +
		authURL + "\n\n" +
		"")

	if openBrowser {
		err = browser.OpenURL(authURL)
		if err != nil {
			return "", err
		}
	}

	return state, nil
}

// ActivateCmd returns a cobra.Command to activate Pachyderm's auth system
func ActivateCmd() *cobra.Command {
	var supplyRootToken bool
	activate := &cobra.Command{
		Short: "Activate Pachyderm's auth system",
		Long: `
Activate Pachyderm's auth system, and restrict access to existing data to the root user`[1:],
		Run: cmdutil.Run(func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var rootToken string
			if supplyRootToken {
				rootToken, err = cmdutil.ReadPassword("")
				if err != nil {
					return errors.Wrapf(err, "error reading token")
				}
			}

			resp, err := c.Activate(c.Ctx(), &auth.ActivateRequest{
				RootToken: strings.TrimSpace(rootToken),
			})
			if err != nil && !auth.IsErrAlreadyActivated(err) {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "error activating Pachyderm auth")
			}

			// If auth is already activated we won't get back a root token,
			// retry activating auth in PPS with the currently configured token
			if !auth.IsErrAlreadyActivated(err) {
				if err := config.WritePachTokenToConfig(resp.PachToken); err != nil {
					return err
				}

				fmt.Println("WARNING: DO NOT LOSE THE AUTH TOKEN BELOW. STORE IT SECURELY FOR THE LIFE OF THE CLUSTER." +
					"THIS TOKEN WILL ALWAYS HAVE ADMIN ACCESS TO FIX THE CLUSTER CONFIGURATION.")
				fmt.Printf("Pachyderm root token:\n%s\n", resp.PachToken)

				c.SetAuthToken(resp.PachToken)
			}

			if _, err := c.ActivateAuth(c.Ctx(), &pps.ActivateAuthRequest{}); err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "error configuring auth for existing PPS pipelines - run `pachctl auth activate` again")
			}
			return nil
		}),
	}
	activate.PersistentFlags().BoolVar(&supplyRootToken, "supply-root-token", false, `
Prompt the user to input a root token on stdin, rather than generating a random one.`[1:])

	return cmdutil.CreateAlias(activate, "auth activate")
}

// DeactivateCmd returns a cobra.Command to delete all ACLs, tokens, and admins,
// deactivating Pachyderm's auth system
func DeactivateCmd() *cobra.Command {
	deactivate := &cobra.Command{
		Short: "Delete all ACLs, tokens, and admins, and deactivate Pachyderm auth",
		Long: "Deactivate Pachyderm's auth system, which will delete ALL auth " +
			"tokens, ACLs and admins, and expose all data in the cluster to any " +
			"user with cluster access. Use with caution.",
		Run: cmdutil.Run(func(args []string) error {
			fmt.Println("Are you sure you want to delete ALL auth information " +
				"(ACLs, tokens, and admins) in this cluster, and expose ALL data? yN")
			confirm, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return err
			}
			if !strings.Contains("yY", confirm[:1]) {
				return errors.Errorf("operation aborted")
			}
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			_, err = c.Deactivate(c.Ctx(), &auth.DeactivateRequest{})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(deactivate, "auth deactivate")
}

// LoginCmd returns a cobra.Command to login to a Pachyderm cluster with your
// GitHub account. Any resources that have been restricted to the email address
// registered with your GitHub account will subsequently be accessible.
func LoginCmd() *cobra.Command {
	var noBrowser bool
	login := &cobra.Command{
		Short: "Log in to Pachyderm",
		Long: "Login to Pachyderm. Any resources that have been restricted to " +
			"the account you have with your ID provider (e.g. GitHub, Okta) " +
			"account will subsequently be accessible.",
		Run: cmdutil.Run(func([]string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			// Issue authentication request to Pachyderm and get response
			var resp *auth.AuthenticateResponse
			var authErr error
			if state, err := requestOIDCLogin(c, !noBrowser); err == nil {
				// Exchange OIDC token for Pachyderm token
				fmt.Println("Retrieving Pachyderm token...")
				resp, authErr = c.Authenticate(
					c.Ctx(),
					&auth.AuthenticateRequest{OIDCState: state})
				if authErr != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(authErr),
						"authorization failed (OIDC state token: %q; Pachyderm logs may "+
							"contain more information)",
						// Print state token as it's logged, for easy searching
						fmt.Sprintf("%s.../%d", state[:len(state)/2], len(state)))
				}
			} else {
				return fmt.Errorf("no authentication providers are configured")
			}

			// Write new Pachyderm token to config
			if authErr != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(authErr), "error authenticating with Pachyderm cluster")
			}
			return config.WritePachTokenToConfig(resp.PachToken)
		}),
	}
	login.PersistentFlags().BoolVarP(&noBrowser, "no-browser", "b", false,
		"If set, don't try to open a web browser")
	return cmdutil.CreateAlias(login, "auth login")
}

// LogoutCmd returns a cobra.Command that deletes your local Pachyderm
// credential, logging you out of your cluster. Note that this is not necessary
// to do before logging in as another user, but is useful for testing.
func LogoutCmd() *cobra.Command {
	logout := &cobra.Command{
		Short: "Log out of Pachyderm by deleting your local credential",
		Long: "Log out of Pachyderm by deleting your local credential. Note that " +
			"it's not necessary to log out before logging in with another account " +
			"(simply run 'pachctl auth login' twice) but 'logout' can be useful on " +
			"shared workstations.",
		Run: cmdutil.Run(func([]string) error {
			cfg, err := config.Read(false, false)
			if err != nil {
				return errors.Wrapf(err, "error reading Pachyderm config (for cluster address)")
			}
			_, context, err := cfg.ActiveContext(true)
			if err != nil {
				return errors.Wrapf(err, "error getting the active context")
			}
			context.SessionToken = ""
			return cfg.Write()
		}),
	}
	return cmdutil.CreateAlias(logout, "auth logout")
}

// WhoamiCmd returns a cobra.Command that deletes your local Pachyderm
// credential, logging you out of your cluster. Note that this is not necessary
// to do before logging in as another user, but is useful for testing.
func WhoamiCmd() *cobra.Command {
	whoami := &cobra.Command{
		Short: "Print your Pachyderm identity",
		Long:  "Print your Pachyderm identity.",
		Run: cmdutil.Run(func([]string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			resp, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "error")
			}
			fmt.Printf("You are \"%s\"\n", resp.Username)
			if resp.TTL > 0 {
				fmt.Printf("session expires: %v\n", time.Now().Add(time.Duration(resp.TTL)*time.Second).Format(time.RFC822))
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(whoami, "auth whoami")
}

// GetAuthTokenCmd returns a cobra command that lets a user get a pachyderm
// token on behalf of themselves or another user
func GetAuthTokenCmd() *cobra.Command {
	var quiet bool
	var ttl string
	getAuthToken := &cobra.Command{
		Use: "{{alias}} [username]",
		Short: "Get an auth token that authenticates the holder as \"username\", " +
			"or the currently signed-in user, if no 'username' is provided",
		Long: "Get an auth token that authenticates the holder as \"username\"; " +
			"or the currently signed-in user, if no 'username' is provided. Only " +
			"cluster admins can obtain an auth token on behalf of another user.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			req := &auth.GetAuthTokenRequest{}
			if ttl != "" {
				d, err := time.ParseDuration(ttl)
				if err != nil {
					return errors.Wrapf(err, "could not parse duration %q", ttl)
				}
				req.TTL = int64(d.Seconds())
			}
			if len(args) == 1 {
				req.Subject = args[0]
			}
			resp, err := c.GetAuthToken(c.Ctx(), req)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if quiet {
				fmt.Println(resp.Token)
			} else {
				fmt.Printf("New credentials:\n  Subject: %s\n  Token: %s\n", resp.Subject, resp.Token)
			}
			return nil
		}),
	}
	getAuthToken.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "if "+
		"set, only print the resulting token (if successful). This is useful for "+
		"scripting, as the output can be piped to use-auth-token")
	getAuthToken.PersistentFlags().StringVar(&ttl, "ttl", "", "if set, the "+
		"resulting auth token will have the given lifetime (or the lifetime"+
		"of the caller's current session, whichever is shorter). This flag should "+
		"be a golang duration (e.g. \"30s\" or \"1h2m3s\"). If unset, tokens will "+
		"have a lifetime of 30 days.")
	return cmdutil.CreateAlias(getAuthToken, "auth get-auth-token")
}

// UseAuthTokenCmd returns a cobra command that lets a user get a pachyderm
// token on behalf of themselves or another user
func UseAuthTokenCmd() *cobra.Command {
	useAuthToken := &cobra.Command{
		Short: "Read a Pachyderm auth token from stdin, and write it to the " +
			"current user's Pachyderm config file",
		Long: "Read a Pachyderm auth token from stdin, and write it to the " +
			"current user's Pachyderm config file",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			fmt.Println("Please paste your Pachyderm auth token:")
			token, err := cmdutil.ReadPassword("")
			if err != nil {
				return errors.Wrapf(err, "error reading token")
			}
			config.WritePachTokenToConfig(strings.TrimSpace(token)) // drop trailing newline
			return nil
		}),
	}
	return cmdutil.CreateAlias(useAuthToken, "auth use-auth-token")
}

// Cmds returns a list of cobra commands for authenticating and authorizing
// users in an auth-enabled Pachyderm cluster.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	auth := &cobra.Command{
		Short: "Auth commands manage access to data in a Pachyderm cluster",
		Long:  "Auth commands manage access to data in a Pachyderm cluster",
	}

	commands = append(commands, cmdutil.CreateAlias(auth, "auth"))
	commands = append(commands, ActivateCmd())
	commands = append(commands, DeactivateCmd())
	commands = append(commands, LoginCmd())
	commands = append(commands, LogoutCmd())
	commands = append(commands, WhoamiCmd())
	commands = append(commands, GetAuthTokenCmd())
	commands = append(commands, UseAuthTokenCmd())
	commands = append(commands, GetConfigCmd())
	commands = append(commands, SetConfigCmd())

	return commands
}
