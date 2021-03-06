package cmds

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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

func printRoleBinding(b *auth.RoleBinding) {
	for principal, roles := range b.Entries {
		roleList := make([]string, 0)
		for r := range roles.Roles {
			roleList = append(roleList, r)
		}
		fmt.Printf("%v: %v\n", principal, roleList)
	}
}

// ActivateCmd returns a cobra.Command to activate Pachyderm's auth system
func ActivateCmd() *cobra.Command {
	var supplyRootToken, onlyActivate bool
	var trustedPeers []string
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
			if _, err := c.PfsAPIClient.ActivateAuth(c.Ctx(), &pfs.ActivateAuthRequest{}); err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "error configuring auth for existing PFS repos - run `pachctl auth activate` again")
			}
			if _, err := c.PpsAPIClient.ActivateAuth(c.Ctx(), &pps.ActivateAuthRequest{}); err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "error configuring auth for existing PPS pipelines - run `pachctl auth activate` again")
			}

			// By default, configure pachd as an OIDC client for the embedded Dex server.
			// If this fails users may need to configure the server manually.
			if !onlyActivate {
				if _, err := c.SetIdentityServerConfig(c.Ctx(), &identity.SetIdentityServerConfigRequest{
					Config: &identity.IdentityServerConfig{
						Issuer: "http://localhost:30658/",
					}}); err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to configure identity server issuer")
				}

				oidcClient, err := c.CreateOIDCClient(c.Ctx(), &identity.CreateOIDCClientRequest{
					Client: &identity.OIDCClient{
						Id:           "pachd",
						Name:         "pachd",
						TrustedPeers: trustedPeers,
						RedirectUris: []string{"http://localhost:30657/authorization-code/callback"},
					},
				})
				if err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to configure OIDC client ID")
				}

				if _, err := c.SetConfiguration(c.Ctx(),
					&auth.SetConfigurationRequest{Configuration: &auth.OIDCConfig{
						Issuer:          "http://localhost:30658/",
						ClientID:        "pachd",
						ClientSecret:    oidcClient.Client.Secret,
						RedirectURI:     "http://localhost:30657/authorization-code/callback",
						LocalhostIssuer: true,
					}}); err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to configure OIDC in pachd")
				}
			}
			return nil
		}),
	}
	activate.PersistentFlags().BoolVar(&supplyRootToken, "supply-root-token", false, `
Prompt the user to input a root token on stdin, rather than generating a random one.`[1:])
	activate.PersistentFlags().BoolVar(&onlyActivate, "only-activate", false, "Activate auth without configuring the OIDC service")
	activate.PersistentFlags().StringSliceVar(&trustedPeers, "trusted-peers", []string{}, "Comma-separated list of OIDC client IDs to trust")

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

// GetRobotTokenCmd returns a cobra command that lets a user get a pachyderm
// token on behalf of themselves or another user
func GetRobotTokenCmd() *cobra.Command {
	var quiet bool
	var ttl string
	getAuthToken := &cobra.Command{
		Use:   "{{alias}} [username]",
		Short: "Get an auth token for a robot user with the specified name.",
		Long:  "Get an auth token for a robot user with the specified name.",
		Run: cmdutil.RunBoundedArgs(1, 1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			req := &auth.GetRobotTokenRequest{
				Robot: args[0],
			}
			if ttl != "" {
				d, err := time.ParseDuration(ttl)
				if err != nil {
					return errors.Wrapf(err, "could not parse duration %q", ttl)
				}
				req.TTL = int64(d.Seconds())
			}
			resp, err := c.GetRobotToken(c.Ctx(), req)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if quiet {
				fmt.Println(resp.Token)
			} else {
				fmt.Printf("Token: %s\n", resp.Token)
			}
			return nil
		}),
	}
	getAuthToken.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "if "+
		"set, only print the resulting token (if successful). This is useful for "+
		"scripting, as the output can be piped to use-auth-token")
	getAuthToken.PersistentFlags().StringVar(&ttl, "ttl", "", "if set, the "+
		"resulting auth token will have the given lifetime. If not set, the token does not expire."+
		" This flag should be a golang duration (e.g. \"30s\" or \"1h2m3s\").")
	return cmdutil.CreateAlias(getAuthToken, "auth get-robot-token")
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

// CheckRepoCmd returns a cobra command that sends an "Authorize" RPC to Pachd, to
// determine whether the specified user has access to the specified repo.
func CheckRepoCmd() *cobra.Command {
	check := &cobra.Command{
		Use:   "{{alias}} <permission> <repo>",
		Short: "Check whether you have the specificed permission on 'repo'",
		Long:  "Check whether you have the specificed permission on 'repo'",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			permission, ok := auth.Permission_value[args[0]]
			if !ok {
				return fmt.Errorf("unknown permission %q", args[0])
			}
			repo := args[1]
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			resp, err := c.Authorize(c.Ctx(), &auth.AuthorizeRequest{
				Resource:    &auth.Resource{Type: auth.ResourceType_REPO, Name: repo},
				Permissions: []auth.Permission{auth.Permission(permission)},
			})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Printf("%t\n", resp.Authorized)
			return nil
		}),
	}
	return cmdutil.CreateAlias(check, "auth check repo")
}

// SetRepoRoleBindingCmd returns a cobra command that sets the roles for a user on a resource
func SetRepoRoleBindingCmd() *cobra.Command {
	setScope := &cobra.Command{
		Use:   "{{alias}} <repo> [role1,role2 | none ] <subject>",
		Short: "Set the roles that 'username' has on 'repo'",
		Long:  "Set the roles that 'username' has on 'repo'",
		Run: cmdutil.RunFixedArgs(3, func(args []string) error {
			var roles []string
			if args[1] == "none" {
				roles = []string{}
			} else {
				roles = strings.Split(args[1], ",")
			}

			subject, repo := args[2], args[0]
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			err = c.ModifyRepoRoleBinding(repo, subject, roles)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(setScope, "auth set repo")
}

// GetRepoRoleBindingCmd returns a cobra command that gets the role bindings for a resource
func GetRepoRoleBindingCmd() *cobra.Command {
	get := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Get the role bindings for 'repo'",
		Long:  "Get the role bindings for 'repo'",
		Run: cmdutil.RunBoundedArgs(1, 1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			repo := args[0]
			resp, err := c.GetRepoRoleBinding(repo)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			printRoleBinding(resp)
			return nil
		}),
	}
	return cmdutil.CreateAlias(get, "auth get repo")
}

// SetClusterRoleBindingCmd returns a cobra command that sets the roles for a user on a resource
func SetClusterRoleBindingCmd() *cobra.Command {
	setScope := &cobra.Command{
		Use:   "{{alias}} [role1,role2 | none ] subject",
		Short: "Set the roles that 'username' has on the cluster",
		Long:  "Set the roles that 'username' has on the cluster",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			var roles []string
			if args[0] == "none" {
				roles = []string{}
			} else {
				roles = strings.Split(args[0], ",")
			}

			subject := args[1]
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			err = c.ModifyClusterRoleBinding(subject, roles)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(setScope, "auth set cluster")
}

// GetClusterRoleBindingCmd returns a cobra command that gets the role bindings for a resource
func GetClusterRoleBindingCmd() *cobra.Command {
	get := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Get the role bindings for 'repo'",
		Long:  "Get the role bindings for 'repo'",
		Run: cmdutil.RunBoundedArgs(0, 0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			resp, err := c.GetClusterRoleBinding()
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			printRoleBinding(resp)
			return nil
		}),
	}
	return cmdutil.CreateAlias(get, "auth get cluster")
}

// Cmds returns a list of cobra commands for authenticating and authorizing
// users in an auth-enabled Pachyderm cluster.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	auth := &cobra.Command{
		Short: "Auth commands manage access to data in a Pachyderm cluster",
		Long:  "Auth commands manage access to data in a Pachyderm cluster",
	}

	get := &cobra.Command{
		Short: "Get the role bindings for a resource",
		Long:  "Get the role bindings for a resource",
	}

	set := &cobra.Command{
		Short: "Set the role bindings for a resource",
		Long:  "Set the role bindings for a resource",
	}

	check := &cobra.Command{
		Short: "Check whether a subject has a permission on a resource",
		Long:  "Check whether a subject has a permission on a resource",
	}

	commands = append(commands, cmdutil.CreateAlias(auth, "auth"))
	commands = append(commands, cmdutil.CreateAlias(get, "auth get"))
	commands = append(commands, cmdutil.CreateAlias(set, "auth set"))
	commands = append(commands, cmdutil.CreateAlias(check, "auth check"))
	commands = append(commands, ActivateCmd())
	commands = append(commands, DeactivateCmd())
	commands = append(commands, LoginCmd())
	commands = append(commands, LogoutCmd())
	commands = append(commands, WhoamiCmd())
	commands = append(commands, GetAuthTokenCmd())
	commands = append(commands, GetRobotTokenCmd())
	commands = append(commands, UseAuthTokenCmd())
	commands = append(commands, GetConfigCmd())
	commands = append(commands, SetConfigCmd())
	commands = append(commands, CheckRepoCmd())
	commands = append(commands, GetRepoRoleBindingCmd())
	commands = append(commands, SetRepoRoleBindingCmd())
	commands = append(commands, GetClusterRoleBindingCmd())
	commands = append(commands, SetClusterRoleBindingCmd())
	return commands
}
