package cmds

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

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
		return "", fmt.Errorf("error reading token: %v", err)
	}
	return strings.TrimSpace(token), nil // drop trailing newline
}

func writePachTokenToCfg(token string) error {
	cfg, err := config.Read()
	if err != nil {
		return fmt.Errorf("error reading Pachyderm config (for cluster "+
			"address): %v", err)
	}
	if cfg.V1 == nil {
		cfg.V1 = &config.ConfigV1{}
	}
	cfg.V1.SessionToken = token
	return cfg.Write()
}

// ActivateCmd returns a cobra.Command to activate Pachyderm's auth system
func ActivateCmd() *cobra.Command {
	activate := &cobra.Command{
		Use:   "activate",
		Short: "Activate Pachyderm's auth system",
		Long:  "Activate Pachyderm's auth system, and restrict access to existing data to cluster admins",
		Run: cmdutil.Run(func(args []string) error {
			token, err := githubLogin()
			if err != nil {
				return err
			}
			fmt.Println("Retrieving Pachyderm token...")

			// Exchange GitHub token for Pachyderm token
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
			resp, err := c.Activate(
				c.Ctx(),
				&auth.ActivateRequest{GitHubToken: token})
			if err != nil {
				return fmt.Errorf("error activating Pachyderm auth: %v",
					grpcutil.ScrubGRPC(err))
			}
			return writePachTokenToCfg(resp.PachToken)
		}),
	}
	return activate
}

// DeactivateCmd returns a cobra.Command to delete all ACLs, tokens, and admins,
// deactivating Pachyderm's auth system
func DeactivateCmd() *cobra.Command {
	deactivate := &cobra.Command{
		Use:   "deactivate",
		Short: "Delete all ACLs, tokens, and admins, and deactivate Pachyderm auth",
		Long: "Deactivate Pachyderm's auth system, which will delete ALL auth " +
			"tokens, ACLs and admins, and expose all data in the cluster to any " +
			"user with cluster access. Use with caution.",
		Run: cmdutil.Run(func(args []string) error {
			fmt.Println("Are you sure you want to delete ALL auth information " +
				"(ACLs, tokens, and admins) in this cluster, and expose ALL data? yN")
			confirm, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if !strings.Contains("yY", confirm[:1]) {
				return fmt.Errorf("operation aborted")
			}
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
			_, err = c.Deactivate(c.Ctx(), &auth.DeactivateRequest{})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return deactivate
}

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
			token, err := githubLogin()
			if err != nil {
				return err
			}
			fmt.Println("Retrieving Pachyderm token...")

			// Exchange GitHub token for Pachyderm token
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
			resp, err := c.Authenticate(
				c.Ctx(),
				&auth.AuthenticateRequest{GitHubToken: token})
			if err != nil {
				return fmt.Errorf("error authenticating with Pachyderm cluster: %v",
					grpcutil.ScrubGRPC(err))
			}
			return writePachTokenToCfg(resp.PachToken)
		}),
	}
	return login
}

// LogoutCmd returns a cobra.Command that deletes your local Pachyderm
// credential, logging you out of your cluster. Note that this is not necessary
// to do before logging in as another user, but is useful for testing.
func LogoutCmd() *cobra.Command {
	logout := &cobra.Command{
		Use:   "logout",
		Short: "Log out of Pachyderm by deleting your local credential",
		Long: "Log out of Pachyderm by deleting your local credential. Note that " +
			"it's not necessary to log out before logging in with another account " +
			"(simply run 'pachctl auth login' twice) but 'logout' can be useful on " +
			"shared workstations.",
		Run: cmdutil.Run(func([]string) error {
			cfg, err := config.Read()
			if err != nil {
				return fmt.Errorf("error reading Pachyderm config (for cluster "+
					"address): %v", err)
			}
			if cfg.V1 == nil {
				return nil
			}
			cfg.V1.SessionToken = ""
			return cfg.Write()
		}),
	}
	return logout
}

// WhoamiCmd returns a cobra.Command that deletes your local Pachyderm
// credential, logging you out of your cluster. Note that this is not necessary
// to do before logging in as another user, but is useful for testing.
func WhoamiCmd() *cobra.Command {
	whoami := &cobra.Command{
		Use:   "whoami",
		Short: "Print your Pachyderm identity",
		Long:  "Print your Pachyderm identity.",
		Run: cmdutil.Run(func([]string) error {
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
			resp, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
			if err != nil {
				return fmt.Errorf("error: %v", grpcutil.ScrubGRPC(err))
			}
			fmt.Printf("You are \"%s\"\n", resp.Username)
			return nil
		}),
	}
	return whoami
}

// CheckCmd returns a cobra command that sends an "Authorize" RPC to Pachd, to
// determine whether the specified user has access to the specified repo.
func CheckCmd() *cobra.Command {
	check := &cobra.Command{
		Use:   "check (none|reader|writer|owner) repo",
		Short: "Check whether you have reader/writer/etc-level access to 'repo'",
		Long: "Check whether you have reader/writer/etc-level access to 'repo'. " +
			"For example, 'pachctl auth check reader private-data' prints \"true\" " +
			"if the you have at least \"reader\" access to the repo " +
			"\"private-data\" (you could be a reader, writer, or owner). Unlike " +
			"`pachctl get-acl`, you do not need to have access to 'repo' to " +
			"discover your own acess level.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			scope, err := auth.ParseScope(args[0])
			if err != nil {
				return err
			}
			repo := args[1]
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
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
	return check
}

// GetCmd returns a cobra command that gets either the ACL for a Pachyderm
// repo or another user's scope of access to that repo
func GetCmd() *cobra.Command {
	setScope := &cobra.Command{
		Use:   "get [username] repo",
		Short: "Get the ACL for 'repo' or the access that 'username' has to 'repo'",
		Long: "Get the ACL for 'repo' or the access that 'username' has to " +
			"'repo'. For example, 'pachctl auth get github-alice private-data' " +
			"prints \"reader\", \"writer\", \"owner\", or \"none\", depending on " +
			"the privileges that \"github-alice\" has in \"repo\". Currently all " +
			"Pachyderm authentication uses GitHub OAuth, so 'username' must be a " +
			"GitHub username",
		Run: cmdutil.RunBoundedArgs(1, 2, func(args []string) error {
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
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
	return setScope
}

// SetScopeCmd returns a cobra command that lets a user set the level of access
// that another user has to a repo
func SetScopeCmd() *cobra.Command {
	setScope := &cobra.Command{
		Use:   "set username (none|reader|writer|owner) repo",
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
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
			_, err = c.SetScope(c.Ctx(), &auth.SetScopeRequest{
				Repo:     repo,
				Scope:    scope,
				Username: username,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return setScope
}

// ListAdminsCmd returns a cobra command that lists the current cluster admins
func ListAdminsCmd() *cobra.Command {
	listAdmins := &cobra.Command{
		Use:   "list-admins",
		Short: "List the current cluster admins",
		Long:  "List the current cluster admins",
		Run: cmdutil.Run(func([]string) error {
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return err
			}
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
	return listAdmins
}

// ModifyAdminsCmd returns a cobra command that modifies the set of current
// cluster admins
func ModifyAdminsCmd() *cobra.Command {
	var add []string
	var remove []string
	modifyAdmins := &cobra.Command{
		Use:   "modify-admins",
		Short: "Modify the current cluster admins",
		Long: "Modify the current cluster admins. --add accepts a comma-" +
			"separated list of users to grant admin status, and --remove accepts a " +
			"comma-separated list of users to revoke admin status",
		Run: cmdutil.Run(func([]string) error {
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return err
			}
			_, err = c.ModifyAdmins(c.Ctx(), &auth.ModifyAdminsRequest{
				Add:    add,
				Remove: remove,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	modifyAdmins.PersistentFlags().StringSliceVar(&add, "add", []string{},
		"Comma-separated list of users to grant admin status")
	modifyAdmins.PersistentFlags().StringSliceVar(&remove, "remove", []string{},
		"Comma-separated list of users revoke admin status")
	return modifyAdmins
}

// Cmds returns a list of cobra commands for authenticating and authorizing
// users in an auth-enabled Pachyderm cluster.
func Cmds() []*cobra.Command {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Auth commands manage access to data in a Pachyderm cluster",
		Long:  "Auth commands manage access to data in a Pachyderm cluster",
	}
	auth.AddCommand(ActivateCmd())
	auth.AddCommand(DeactivateCmd())
	auth.AddCommand(LoginCmd())
	auth.AddCommand(LogoutCmd())
	auth.AddCommand(WhoamiCmd())
	auth.AddCommand(CheckCmd())
	auth.AddCommand(SetScopeCmd())
	auth.AddCommand(GetCmd())
	auth.AddCommand(ListAdminsCmd())
	auth.AddCommand(ModifyAdminsCmd())
	return []*cobra.Command{auth}
}
