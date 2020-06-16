package cmds

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/auth/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pkg/browser"

	"github.com/spf13/cobra"
)

var githubAuthLink = `https://github.com/login/oauth/authorize?client_id=d3481e92b4f09ea74ff8&redirect_uri=https%3A%2F%2Fpachyderm.io%2Flogin-hook%2Fdisplay-token.html`

func githubLogin() (string, error) {
	fmt.Println("(1) Please paste this link into a browser:\n\n" +
		githubAuthLink + "\n\n" +
		"(You will be directed to GitHub and asked to authorize Pachyderm's " +
		"login app on GitHub. If you accept, you will be given a token to " +
		"paste here, which will give you an externally verified account in " +
		"this Pachyderm cluster)\n\n(2) Please paste the token you receive " +
		"from GitHub here:")
	token, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", errors.Wrapf(err, "error reading token")
	}
	return strings.TrimSpace(token), nil // drop trailing newline
}

func requestOIDCLogin(c *client.APIClient) (string, error) {
	state := server.CryptoString(30)

	var authURL string
	loginInfo, err := c.GetOIDCLogin(c.Ctx(), &auth.GetOIDCLoginRequest{State: state})
	if err != nil {
		return "", err
	}
	authURL = loginInfo.LoginURL

	// print the prepared URL and promp the user to click on it
	fmt.Println("(1) Please paste this link into a browser:\n\n" +
		authURL + "\n\n" +
		"(You will be directed to your IdP and asked to authorize Pachyderm's " +
		"login app on your IdP.)")

	err = browser.OpenURL(authURL)
	if err != nil {
		return "", err
	}
	// receive token
	authError, err := c.GetOIDCError(c.Ctx(), &auth.GetOIDCErrorRequest{})
	if err != nil {
		return "", err
	}

	if authError != nil {
		if authError.Error != "" {
			fmt.Printf("authorization error: %v\n", authError.Error)
		}
	}

	fmt.Println("Authorized!")

	return state, nil // drop trailing newline
}

func writePachTokenToCfg(token string) error {
	cfg, err := config.Read(false)
	if err != nil {
		return errors.Wrapf(err, "error reading Pachyderm config (for cluster address)")
	}
	_, context, err := cfg.ActiveContext(true)
	if err != nil {
		return errors.Wrapf(err, "error getting the active context")
	}
	context.SessionToken = token
	if err := cfg.Write(); err != nil {
		return errors.Wrapf(err, "error writing pachyderm config")
	}
	return nil
}

// ActivateCmd returns a cobra.Command to activate Pachyderm's auth system
func ActivateCmd() *cobra.Command {
	var initialAdmin string
	var url string
	activate := &cobra.Command{
		Short: "Activate Pachyderm's auth system",
		Long: `
Activate Pachyderm's auth system, and restrict access to existing data to the
user running the command (or the argument to --initial-admin), who will be the
first cluster admin`[1:],
		Run: cmdutil.Run(func(args []string) error {
			// var state string
			var err error

			// Exchange GitHub token for Pachyderm token
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			if !strings.HasPrefix(initialAdmin, auth.RobotPrefix) {

				// 	// token, err = githubLogin()
				// 	_, err = requestOIDCLogin(c, "", clientID, clientSecret, email)
				// 	if err != nil {

				// 		return err
				// 	}

				// }
			}

			fmt.Println("Retrieving Pachyderm token...")

			resp, err := c.Activate(c.Ctx(),
				&auth.ActivateRequest{
					// GitHubToken: state,
					Subject: initialAdmin,
				})
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "error activating Pachyderm auth")
			}

			if err := writePachTokenToCfg(resp.PachToken); err != nil {
				return err
			}
			if strings.HasPrefix(initialAdmin, auth.RobotPrefix) {
				fmt.Println("WARNING: DO NOT LOSE THE ROBOT TOKEN BELOW WITHOUT " +
					"ADDING OTHER ADMINS.\nIF YOU DO, YOU WILL BE PERMANENTLY LOCKED OUT " +
					"OF YOUR CLUSTER!")
				fmt.Printf("Pachyderm token for \"%s\":\n%s\n", initialAdmin, resp.PachToken)
			}
			return nil
		}),
	}
	activate.PersistentFlags().StringVar(&initialAdmin, "initial-admin", "", `
The subject (robot user or github user) who
will be the first cluster admin; the user running 'activate' will identify as
this user once auth is active.  If you set 'initial-admin' to a robot
user, pachctl will print that robot user's Pachyderm token; this token is
effectively a root token, and if it's lost you will be locked out of your
cluster`[1:])
	activate.PersistentFlags().StringVar(&url, "url", "u", `Base URL for your identity provider (used for Open ID Connect flow)`)
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
	var useOTP bool
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
			if useOTP {
				// Exhange short-lived Pachyderm auth code for long-lived Pachyderm token
				fmt.Println("Please enter your Pachyderm One-Time Password:")
				code, err := bufio.NewReader(os.Stdin).ReadString('\n')
				if err != nil {
					return errors.Wrapf(err, "error reading One-Time Password")
				}
				code = strings.TrimSpace(code) // drop trailing newline
				resp, authErr = c.Authenticate(
					c.Ctx(),
					&auth.AuthenticateRequest{OneTimePassword: code})
				// } else if github {
				// 	// Exchange GitHub token for Pachyderm token
				// 	token, err := githubLogin()
				// 	if err != nil {
				// 		return err
				// 	}
				// 	fmt.Println("Retrieving Pachyderm token...")
				// 	resp, authErr = c.Authenticate(
				// 		c.Ctx(),
				// 		&auth.AuthenticateRequest{GitHubToken: token})
			} else {
				// Exchange OIDC token for Pachyderm token
				state, err := requestOIDCLogin(c)
				if err != nil {
					return err
				}
				fmt.Println("Retrieving Pachyderm token...")
				resp, authErr = c.Authenticate(
					c.Ctx(),
					&auth.AuthenticateRequest{OIDCState: state})
			}

			// Write new Pachyderm token to config
			if authErr != nil {
				if auth.IsErrPartiallyActivated(authErr) {
					return errors.Wrapf(authErr, "if pachyderm is stuck in this state, you "+
						"can revert by running 'pachctl auth deactivate' or retry by "+
						"running 'pachctl auth activate' again")
				}
				return errors.Wrapf(grpcutil.ScrubGRPC(authErr), "error authenticating with Pachyderm cluster")
			}
			return writePachTokenToCfg(resp.PachToken)
		}),
	}
	login.PersistentFlags().BoolVarP(&useOTP, "one-time-password", "o", false,
		"If set, authenticate with a Dash-provided One-Time Password, rather than "+
			"via GitHub")
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
			cfg, err := config.Read(false)
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
			if resp.IsAdmin {
				fmt.Println("You are an administrator of this Pachyderm cluster")
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(whoami, "auth whoami")
}

// CheckCmd returns a cobra command that sends an "Authorize" RPC to Pachd, to
// determine whether the specified user has access to the specified repo.
func CheckCmd() *cobra.Command {
	check := &cobra.Command{
		Use:   "{{alias}} (none|reader|writer|owner) <repo>",
		Short: "Check whether you have reader/writer/etc-level access to 'repo'",
		Long: "Check whether you have reader/writer/etc-level access to 'repo'. " +
			"For example, 'pachctl auth check reader private-data' prints \"true\" " +
			"if the you have at least \"reader\" access to the repo " +
			"\"private-data\" (you could be a reader, writer, or owner). Unlike " +
			"`pachctl auth get`, you do not need to have access to 'repo' to " +
			"discover your own access level.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			scope, err := auth.ParseScope(args[0])
			if err != nil {
				return err
			}
			repo := args[1]
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			resp, err := c.Authorize(c.Ctx(), &auth.AuthorizeRequest{
				Repo:  repo,
				Scope: scope,
			})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Printf("%t\n", resp.Authorized)
			return nil
		}),
	}
	return cmdutil.CreateAlias(check, "auth check")
}

// GetCmd returns a cobra command that gets either the ACL for a Pachyderm
// repo or another user's scope of access to that repo
func GetCmd() *cobra.Command {
	get := &cobra.Command{
		Use:   "{{alias}} [<username>] <repo>",
		Short: "Get the ACL for 'repo' or the access that 'username' has to 'repo'",
		Long: "Get the ACL for 'repo' or the access that 'username' has to " +
			"'repo'. For example, 'pachctl auth get github-alice private-data' " +
			"prints \"reader\", \"writer\", \"owner\", or \"none\", depending on " +
			"the privileges that \"github-alice\" has in \"repo\". Currently all " +
			"Pachyderm authentication uses GitHub OAuth, so 'username' must be a " +
			"GitHub username",
		Run: cmdutil.RunBoundedArgs(1, 2, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			if len(args) == 1 {
				// Get ACL for a repo
				repo := args[0]
				resp, err := c.GetACL(c.Ctx(), &auth.GetACLRequest{
					Repo: repo,
				})
				if err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				t := template.Must(template.New("ACLEntries").Parse(
					"{{range .}}{{.Username }}: {{.Scope}}\n{{end}}"))
				return t.Execute(os.Stdout, resp.Entries)
			}
			// Get User's scope on an acl
			username, repo := args[0], args[1]
			resp, err := c.GetScope(c.Ctx(), &auth.GetScopeRequest{
				Repos:    []string{repo},
				Username: username,
			})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Println(resp.Scopes[0].String())
			return nil
		}),
	}
	return cmdutil.CreateAlias(get, "auth get")
}

// SetScopeCmd returns a cobra command that lets a user set the level of access
// that another user has to a repo
func SetScopeCmd() *cobra.Command {
	setScope := &cobra.Command{
		Use:   "{{alias}} <username> (none|reader|writer|owner) <repo>",
		Short: "Set the scope of access that 'username' has to 'repo'",
		Long: "Set the scope of access that 'username' has to 'repo'. For " +
			"example, 'pachctl auth set github-alice none private-data' prevents " +
			"\"github-alice\" from interacting with the \"private-data\" repo in any " +
			"way (the default). Similarly, 'pachctl auth set github-alice reader " +
			"private-data' would let \"github-alice\" read from \"private-data\" but " +
			"not create commits (writer) or modify the repo's access permissions " +
			"(owner). Currently all Pachyderm authentication uses GitHub OAuth, so " +
			"'username' must be a GitHub username",
		Run: cmdutil.RunFixedArgs(3, func(args []string) error {
			scope, err := auth.ParseScope(args[1])
			if err != nil {
				return err
			}
			username, repo := args[0], args[2]
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			_, err = c.SetScope(c.Ctx(), &auth.SetScopeRequest{
				Repo:     repo,
				Scope:    scope,
				Username: username,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(setScope, "auth set")
}

// ListAdminsCmd returns a cobra command that lists the current cluster admins
func ListAdminsCmd() *cobra.Command {
	listAdmins := &cobra.Command{
		Short: "List the current cluster admins",
		Long:  "List the current cluster admins",
		Run: cmdutil.Run(func([]string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()
			resp, err := c.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			for _, user := range resp.Admins {
				fmt.Println(user)
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(listAdmins, "auth list-admins")
}

// ModifyAdminsCmd returns a cobra command that modifies the set of current
// cluster admins
func ModifyAdminsCmd() *cobra.Command {
	var add []string
	var remove []string
	modifyAdmins := &cobra.Command{
		Short: "Modify the current cluster admins",
		Long: "Modify the current cluster admins. --add accepts a comma-" +
			"separated list of users to grant admin status, and --remove accepts a " +
			"comma-separated list of users to revoke admin status",
		Run: cmdutil.Run(func([]string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()
			_, err = c.ModifyAdmins(c.Ctx(), &auth.ModifyAdminsRequest{
				Add:    add,
				Remove: remove,
			})
			if auth.IsErrPartiallyActivated(err) {
				return errors.Wrapf(err, "Errored, if pachyderm is stuck in this state, you "+
					"can revert by running 'pachctl auth deactivate' or retry by "+
					"running 'pachctl auth activate' again")
			}
			return grpcutil.ScrubGRPC(err)
		}),
	}
	modifyAdmins.PersistentFlags().StringSliceVar(&add, "add", []string{},
		"Comma-separated list of users to grant admin status")
	modifyAdmins.PersistentFlags().StringSliceVar(&remove, "remove", []string{},
		"Comma-separated list of users revoke admin status")
	return cmdutil.CreateAlias(modifyAdmins, "auth modify-admins")
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
			token, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return errors.Wrapf(err, "error reading token")
			}
			writePachTokenToCfg(strings.TrimSpace(token)) // drop trailing newline
			return nil
		}),
	}
	return cmdutil.CreateAlias(useAuthToken, "auth use-auth-token")
}

// GetOneTimePasswordCmd returns a cobra command that lets a user get an OTP.
func GetOneTimePasswordCmd() *cobra.Command {
	var ttl string
	getOneTimePassword := &cobra.Command{
		Use: "{{alias}} <username>",
		Short: "Get a one-time password that authenticates the holder as " +
			"\"username\", or the currently signed in user if no 'username' is " +
			"specified",
		Long: "Get a one-time password that authenticates the holder as " +
			"\"username\", or the currently signed in user if no 'username' is " +
			"specified. Only cluster admins may obtain a one-time password on " +
			"behalf of another user.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			req := &auth.GetOneTimePasswordRequest{}
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
			resp, err := c.GetOneTimePassword(c.Ctx(), req)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Println(resp.Code)
			return nil
		}),
	}
	getOneTimePassword.PersistentFlags().StringVar(&ttl, "ttl", "", "if set, "+
		"the resulting one-time password will have the given lifetime (or the "+
		"lifetime of the caller's current session, whichever is shorter). This "+
		"flag should be a golang duration (e.g. \"30s\" or \"1h2m3s\"). If unset, "+
		"one-time passwords will have a lifetime of 5 minutes")
	return cmdutil.CreateAlias(getOneTimePassword, "auth get-otp")
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
	commands = append(commands, CheckCmd())
	commands = append(commands, SetScopeCmd())
	commands = append(commands, GetCmd())
	commands = append(commands, ListAdminsCmd())
	commands = append(commands, ModifyAdminsCmd())
	commands = append(commands, GetAuthTokenCmd())
	commands = append(commands, UseAuthTokenCmd())
	commands = append(commands, GetConfigCmd())
	commands = append(commands, SetConfigCmd())
	commands = append(commands, GetOneTimePasswordCmd())

	return commands
}
