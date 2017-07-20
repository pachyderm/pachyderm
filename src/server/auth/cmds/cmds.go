package cmds

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra"
)

var githubAuthLink = `https://github.com/login/oauth/authorize?client_id=d3481e92b4f09ea74ff8&redirect_uri=https%3A%2F%2Fpachyderm.io%2Flogin-hook%2Fdisplay-token.html`

// ActivateCmd returns a cobra.Command to activate the security features of
// Pachyderm within a Pachyderm cluster. All repos will go from
// publicly-accessible to accessible only by the owner, who can subsequently add
// users
func ActivateCmd() *cobra.Command {
	var admins []string
	activate := &cobra.Command{
		Use: "activate activation-code",
		Short: "Activate the security features of pachyderm with an activation " +
			"code",
		Long: "Activate the security features of pachyderm with an activation " +
			"code",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			activationCode := args[0]
			if len(admins) == 0 {
				return fmt.Errorf("must specify at least one cluster admin to enable " +
					"auth")
			}
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %s", err.Error())
			}
			_, err = c.AuthAPIClient.Activate(
				c.Ctx(),
				&auth.ActivateRequest{
					ActivationCode: activationCode,
					Admins:         admins,
				})
			return err
		}),
	}
	activate.PersistentFlags().StringSliceVar(&admins, "admins", []string{},
		"Comma-separated list of users who will be cluster admins once security "+
			"is enabled. This list cannot be empty, as only admins can appoint new "+
			"admins, so if a cluster has no admins to begin with, none can be "+
			"appointed.")
	return activate
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
			fmt.Println("(1) Please paste this link into a browser:\n\n" +
				githubAuthLink + "\n\n" +
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
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %s", err.Error())
			}
			resp, err := c.AuthAPIClient.Authenticate(
				c.Ctx(),
				&auth.AuthenticateRequest{GithubToken: token})
			if err != nil {
				return fmt.Errorf("error authenticating with Pachyderm cluster: %s",
					err.Error())
			}
			if cfg.V1 == nil {
				cfg.V1 = &config.ConfigV1{}
			}
			cfg.V1.SessionToken = resp.PachToken
			return cfg.Write()
		}),
	}
	return login
}

// CheckCmd returns a cobra command that sends an "Authorize" RPC to Pachd, to
// determine whether the specified user has access to the specified repo.
func CheckCmd() *cobra.Command {
	check := &cobra.Command{
		Use:   "check (none|reader|writer|owner) repo",
		Short: "Check whether you have reader/writer/etc-level access to 'repo'",
		Long: "Check whether you or another user has a reader/writer/etc-level " +
			"access to 'repo'. For " +
			"example, 'pachctl auth check reader private-data' prints " +
			"\"true\" if the you have at least \"reader\" access " +
			"to the repo \"private-data\" (you could be a reader, writer, " +
			"or owner). Unlike. `pachctl get-acl`, you do not need to have access " +
			"to 'repo' to discover your own acess level.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			scope, err := auth.ParseScope(args[0])
			if err != nil {
				return err
			}
			repo := args[1]
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %s", err.Error())
			}
			resp, err := c.AuthAPIClient.Authorize(
				c.Ctx(),
				&auth.AuthorizeRequest{
					Repo:  &pfs.Repo{Name: repo},
					Scope: scope,
				})
			if err != nil {
				return err
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
				return fmt.Errorf("could not connect: %s", err.Error())
			}
			if len(args) == 1 {
				// Get ACL for a repo
				repo := args[0]
				resp, err := c.AuthAPIClient.GetACL(
					c.Ctx(),
					&auth.GetACLRequest{
						Repo: &pfs.Repo{Name: repo},
					})
				if err != nil {
					return err
				}
				fmt.Println(resp.Acl.String())
				return nil
			}
			// Get User's scope on an acl
			username, repo := args[0], args[1]
			resp, err := c.AuthAPIClient.GetScope(
				c.Ctx(),
				&auth.GetScopeRequest{
					Repo:     &pfs.Repo{Name: repo},
					Username: username,
				})
			if err != nil {
				return err
			}
			fmt.Println(resp.Scope.String())
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
				return fmt.Errorf("could not connect: %s", err.Error())
			}
			_, err = c.AuthAPIClient.SetScope(
				c.Ctx(),
				&auth.SetScopeRequest{
					Repo:     &pfs.Repo{Name: repo},
					Scope:    scope,
					Username: username,
				})
			return err
		}),
	}
	return setScope
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
	auth.AddCommand(CheckCmd())
	auth.AddCommand(SetScopeCmd())
	auth.AddCommand(GetCmd())
	return []*cobra.Command{LoginCmd(), auth}
}
