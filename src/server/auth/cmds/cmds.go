package cmds

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pkg/browser"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/spf13/cobra"
)

func requestOIDCLogin(c *client.APIClient, openBrowser bool) (string, error) {
	var authURL string
	loginInfo, err := c.GetOIDCLogin(c.Ctx(), &auth.GetOIDCLoginRequest{})
	if err != nil {
		return "", err
	}
	authURL = loginInfo.LoginUrl
	state := loginInfo.State

	// print the prepared URL and promp the user to click on it
	fmt.Println("You will momentarily be directed to your IdP and asked to authorize Pachyderm's " +
		"login app on your IdP.\n\nPaste the following URL into a browser if not automatically redirected:\n\n" +
		authURL + "\n\n" +
		"")

	if openBrowser {
		if browser.OpenURL(authURL) != nil {
			fmt.Println("Couldn't open a browser, visit the page manually.")
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
func ActivateCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var enterprise, supplyRootToken, onlyActivate bool
	var issuer, redirect, clientId string
	var trustedPeers, scopes []string
	activate := &cobra.Command{
		Short: "Activate Pachyderm's auth system",
		Long: `
Activate Pachyderm's auth system, and restrict access to existing data to the root user`[1:],
		Run: cmdutil.Run(func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, enterprise)
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
				if err := config.WritePachTokenToConfig(resp.PachToken, enterprise); err != nil {
					return err
				}

				fmt.Println("WARNING: DO NOT LOSE THE AUTH TOKEN BELOW. STORE IT SECURELY FOR THE LIFE OF THE CLUSTER." +
					"THIS TOKEN WILL ALWAYS HAVE ADMIN ACCESS TO FIX THE CLUSTER CONFIGURATION.")
				fmt.Printf("Pachyderm root token:\n%s\n", resp.PachToken)

				c.SetAuthToken(resp.PachToken)
			}

			// The enterprise server doesn't have PFS or PPS enabled
			if !enterprise {
				if _, err := c.PfsAPIClient.ActivateAuth(c.Ctx(), &pfs.ActivateAuthRequest{}); err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "error configuring auth for existing PFS repos - run `pachctl auth activate` again")
				}
				if _, err := c.PpsAPIClient.ActivateAuth(c.Ctx(), &pps.ActivateAuthRequest{}); err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "error configuring auth for existing PPS pipelines - run `pachctl auth activate` again")
				}
			}

			// If the user wants to manually configure OIDC, stop here
			if onlyActivate {
				return nil
			}

			// Check whether the enterprise context is a separate identity server
			conf, err := config.Read(false, true)
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to read configuration file")
			}

			activeContext, _, err := conf.ActiveContext(true)
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to get active context")
			}

			enterpriseContext, _, err := conf.ActiveEnterpriseContext(true)
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to get active context")
			}

			// In a single-node pachd deployment, configure the OIDC server.
			// Otherwise just register with the existing identity server.
			if activeContext == enterpriseContext || enterprise {
				if _, err := c.SetIdentityServerConfig(c.Ctx(), &identity.SetIdentityServerConfigRequest{
					Config: &identity.IdentityServerConfig{
						Issuer: issuer,
					}}); err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to configure identity server issuer")
				}

				oidcClient, err := c.CreateOIDCClient(c.Ctx(), &identity.CreateOIDCClientRequest{
					Client: &identity.OIDCClient{
						Id:           clientId,
						Name:         clientId,
						TrustedPeers: trustedPeers,
						RedirectUris: []string{redirect},
					},
				})
				if err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to configure OIDC client ID")
				}

				if _, err := c.SetConfiguration(c.Ctx(),
					&auth.SetConfigurationRequest{Configuration: &auth.OIDCConfig{
						Issuer:          issuer,
						ClientId:        clientId,
						ClientSecret:    oidcClient.Client.Secret,
						RedirectUri:     redirect,
						LocalhostIssuer: true,
						Scopes:          scopes,
					}}); err != nil {
					err = errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to configure OIDC in pachd")
					_, deleteErr := c.DeleteOIDCClient(c.Ctx(), &identity.DeleteOIDCClientRequest{Id: oidcClient.Client.Id})
					if deleteErr != nil {
						deleteErr = errors.Wrapf(grpcutil.ScrubGRPC(deleteErr), "failed to rollback creation of client with ID: %v."+
							"to retry auth activation, first delete this client with 'pachctl idp delete-client %v'.",
							oidcClient.Client.Id, oidcClient.Client.Id)
						return errors.Wrapf(err, deleteErr.Error())
					}
					return err
				}
			} else {
				ec, err := pachctlCfg.NewOnUserMachine(ctx, true)
				if err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to get enterprise server client")
				}
				idCfg, err := ec.GetIdentityServerConfig(ec.Ctx(), &identity.GetIdentityServerConfigRequest{})
				if err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to get identity server issuer")
				}

				oidcClient, err := ec.CreateOIDCClient(ec.Ctx(), &identity.CreateOIDCClientRequest{
					Client: &identity.OIDCClient{
						Id:           clientId,
						Name:         clientId,
						TrustedPeers: trustedPeers,
						RedirectUris: []string{redirect},
					},
				})
				if err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to configure OIDC client ID")
				}

				if _, err := c.SetConfiguration(c.Ctx(),
					&auth.SetConfigurationRequest{Configuration: &auth.OIDCConfig{
						Issuer:          idCfg.Config.Issuer,
						ClientId:        clientId,
						ClientSecret:    oidcClient.Client.Secret,
						RedirectUri:     redirect,
						LocalhostIssuer: false,
						Scopes:          scopes,
					}}); err != nil {
					err = errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to configure OIDC in pachd.")
					_, deleteErr := c.DeleteOIDCClient(c.Ctx(), &identity.DeleteOIDCClientRequest{Id: oidcClient.Client.Id})
					if deleteErr != nil {
						deleteErr = errors.Wrapf(grpcutil.ScrubGRPC(deleteErr), "failed to rollback creation of client with ID: %v."+
							"to retry auth activation, first delete this client with 'pachctl idp delete-client %v'.",
							oidcClient.Client.Id, oidcClient.Client.Id)
						return errors.Wrapf(err, deleteErr.Error())
					}
					return err
				}
			}
			return nil
		}),
	}
	activate.PersistentFlags().BoolVar(&supplyRootToken, "supply-root-token", false, `
Prompt the user to input a root token on stdin, rather than generating a random one.`[1:])
	activate.PersistentFlags().BoolVar(&onlyActivate, "only-activate", false, "Activate auth without configuring the OIDC service")
	activate.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Activate auth on the active enterprise context")
	activate.PersistentFlags().StringVar(&issuer, "issuer", "http://pachd:1658/", "The issuer for the OIDC service")
	activate.PersistentFlags().StringVar(&redirect, "redirect", "http://localhost:30657/authorization-code/callback", "The redirect URL for the OIDC service")
	activate.PersistentFlags().StringVar(&clientId, "client-id", "pachd", "The client ID for this pachd")
	activate.PersistentFlags().StringSliceVar(&trustedPeers, "trusted-peers", []string{}, "Comma-separated list of OIDC client IDs to trust")
	activate.PersistentFlags().StringSliceVar(&scopes, "scopes", auth.DefaultOIDCScopes, "Comma-separated list of scopes to request")

	return cmdutil.CreateAlias(activate, "auth activate")
}

// DeactivateCmd returns a cobra.Command to delete all ACLs, tokens, and admins,
// deactivating Pachyderm's auth system
func DeactivateCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var enterprise bool
	deactivate := &cobra.Command{
		Short: "Delete all ACLs, tokens, admins, IDP integrations and OIDC clients, and deactivate Pachyderm auth",
		Long: "Deactivate Pachyderm's auth and identity systems, which will delete ALL auth " +
			"tokens, ACLs and admins, IDP integrations and OIDC clients, and expose all data " +
			"in the cluster to any user with cluster access. Use with caution.",
		Run: cmdutil.Run(func(args []string) error {
			fmt.Println("Are you sure you want to delete ALL auth information " +
				"(ACLs, tokens, and admins) in this cluster, and expose ALL data? yN")
			confirm, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return errors.EnsureStack(err)
			}
			if !strings.Contains("yY", confirm[:1]) {
				return errors.Errorf("operation aborted")
			}
			c, err := pachctlCfg.NewOnUserMachine(ctx, enterprise)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			// Delete any data from the identity server
			if _, err := c.IdentityAPIClient.DeleteAll(c.Ctx(), &identity.DeleteAllRequest{}); err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			_, err = c.Deactivate(c.Ctx(), &auth.DeactivateRequest{})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	deactivate.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Deactivate auth on the active enterprise context")
	return cmdutil.CreateAlias(deactivate, "auth deactivate")
}

// LoginCmd returns a cobra.Command to login to a Pachyderm cluster with your
// GitHub account. Any resources that have been restricted to the email address
// registered with your GitHub account will subsequently be accessible.
func LoginCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var noBrowser, enterprise, idToken bool
	login := &cobra.Command{
		Short: "Log in to Pachyderm",
		Long: "Login to Pachyderm. Any resources that have been restricted to " +
			"the account you have with your ID provider (e.g. GitHub, Okta) " +
			"account will subsequently be accessible.",
		Run: cmdutil.Run(func([]string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, enterprise)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			// Issue authentication request to Pachyderm and get response
			var resp *auth.AuthenticateResponse
			var authErr error
			if idToken {
				token, err := cmdutil.ReadPassword("ID token: ")
				if err != nil {
					return errors.Wrapf(err, "could not read id token")
				}
				resp, authErr = c.Authenticate(
					c.Ctx(),
					&auth.AuthenticateRequest{IdToken: strings.TrimSpace(token)})
				if authErr != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(authErr),
						"authorization failed (Pachyderm logs may contain more information)")
				}
			} else {
				if state, err := requestOIDCLogin(c, !noBrowser); err == nil {
					// Exchange OIDC token for Pachyderm token
					fmt.Println("Retrieving Pachyderm token...")
					resp, authErr = c.Authenticate(
						c.Ctx(),
						&auth.AuthenticateRequest{OidcState: state})
					if authErr != nil {
						return errors.Wrapf(grpcutil.ScrubGRPC(authErr),
							"authorization failed (OIDC state token: %q; Pachyderm logs may "+
								"contain more information)",
							// Print state token as it's logged, for easy searching
							fmt.Sprintf("%s.../%d", state[:len(state)/2], len(state)))
					}
				} else {
					return errors.Errorf("no authentication providers are configured")
				}
			}
			if authErr != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(authErr), "error authenticating with Pachyderm cluster")
			}
			return config.WritePachTokenToConfig(resp.PachToken, enterprise)
		}),
	}
	login.PersistentFlags().BoolVarP(&noBrowser, "no-browser", "b", false,
		"If set, don't try to open a web browser")
	login.PersistentFlags().BoolVarP(&idToken, "id-token", "t", false,
		"If set, read an ID token on stdin to authenticate the user")
	login.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Login for the active enterprise context")
	return cmdutil.CreateAlias(login, "auth login")
}

// LogoutCmd returns a cobra.Command that deletes your local Pachyderm
// credential, logging you out of your cluster. Note that this is not necessary
// to do before logging in as another user, but is useful for testing.
func LogoutCmd() *cobra.Command {
	var enterprise bool
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
			if enterprise {
				_, context, err := cfg.ActiveEnterpriseContext(true)
				if err != nil {
					return errors.Wrapf(err, "error getting the active context")
				}
				context.SessionToken = ""
			} else {
				_, context, err := cfg.ActiveContext(true)
				if err != nil {
					return errors.Wrapf(err, "error getting the active context")
				}
				context.SessionToken = ""
			}

			return cfg.Write()
		}),
	}
	logout.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Log out of the active enterprise context")
	return cmdutil.CreateAlias(logout, "auth logout")
}

// WhoamiCmd returns a cobra.Command that deletes your local Pachyderm
// credential, logging you out of your cluster. Note that this is not necessary
// to do before logging in as another user, but is useful for testing.
func WhoamiCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var enterprise bool
	whoami := &cobra.Command{
		Short: "Print your Pachyderm identity",
		Long:  "Print your Pachyderm identity.",
		Run: cmdutil.Run(func([]string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, enterprise)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			resp, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "error")
			}
			fmt.Printf("You are %q\n", resp.Username)
			if e := resp.Expiration; e != nil {
				fmt.Printf("session expires: %v\n", e.AsTime().Format(time.RFC822))
			}
			return nil
		}),
	}
	whoami.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "")
	return cmdutil.CreateAlias(whoami, "auth whoami")
}

// GetRobotTokenCmd returns a cobra command that lets a user get a pachyderm
// token on behalf of themselves or another user
func GetRobotTokenCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var enterprise bool
	var quiet bool
	var ttl string
	getAuthToken := &cobra.Command{
		Use:   "{{alias}} [username]",
		Short: "Get an auth token for a robot user with the specified name.",
		Long:  "Get an auth token for a robot user with the specified name.",
		Run: cmdutil.RunBoundedArgs(1, 1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, enterprise)
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
				req.Ttl = int64(d.Seconds())
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
	getAuthToken.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Get a robot token for the enterprise context")
	return cmdutil.CreateAlias(getAuthToken, "auth get-robot-token")
}

// RevokeCmd returns a cobra.Command that revokes a Pachyderm token.
func RevokeCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var enterprise bool
	var token string
	var user string
	revoke := &cobra.Command{
		Short: "Revoke a Pachyderm auth token",
		Long:  "Revoke a Pachyderm auth token.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			if token == "" && user == "" {
				return errors.Errorf("one of --token or --user must be set")
			} else if token != "" && user != "" {
				return errors.Errorf("only one of --token or --user may be set")
			}

			c, err := pachctlCfg.NewOnUserMachine(ctx, enterprise)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			// handle plurals properly
			_ = message.Set(language.English, "%d auth token(s) revoked", plural.Selectf(1, "",
				plural.One, "1 auth token revoked",
				plural.Other, "%d auth tokens revoked",
			))
			p := message.NewPrinter(language.English)

			if token != "" {
				resp, err := c.RevokeAuthToken(c.Ctx(), &auth.RevokeAuthTokenRequest{Token: token})
				if err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "error")
				}
				fmt.Println(p.Sprintf("%d auth token(s) revoked", resp.Number))
			} else {
				resp, err := c.RevokeAuthTokensForUser(c.Ctx(), &auth.RevokeAuthTokensForUserRequest{Username: user})
				if err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "error")
				}
				fmt.Println(p.Sprintf("%d auth token(s) revoked", resp.Number))
			}
			return nil
		}),
	}
	revoke.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Revoke an auth token (or all auth tokens minted for one user) on the enterprise server")
	revoke.PersistentFlags().StringVar(&token, "token", "", "Pachyderm auth token that should be revoked (one of --token or --user must be set)")
	revoke.PersistentFlags().StringVar(&user, "user", "", "User whose Pachyderm auth tokens should be revoked (one of --token or --user must be set)")
	return cmdutil.CreateAlias(revoke, "auth revoke")
}

func GetGroupsCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var enterprise bool
	getGroups := &cobra.Command{
		Use:   "{{alias}} [username]",
		Short: "Get the list of groups a user belongs to",
		Long:  "Get the list of groups a user belongs to. If no user is specified, the current user's groups are listed.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, enterprise)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var resp *auth.GetGroupsResponse
			if len(args) == 1 {
				resp, err = c.GetGroupsForPrincipal(c.Ctx(), &auth.GetGroupsForPrincipalRequest{Principal: args[0]})
			} else {
				resp, err = c.GetGroups(c.Ctx(), &auth.GetGroupsRequest{})
			}

			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Println(strings.Join(resp.Groups, "\n"))
			return nil
		}),
	}
	getGroups.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Get group membership info from the enterprise server")
	return cmdutil.CreateAlias(getGroups, "auth get-groups")
}

// UseAuthTokenCmd returns a cobra command that lets a user get a pachyderm
// token on behalf of themselves or another user
func UseAuthTokenCmd() *cobra.Command {
	var enterprise bool
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
			return config.WritePachTokenToConfig(strings.TrimSpace(token), enterprise) // drop trailing newline
		}),
	}
	useAuthToken.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Use the token for the enterprise context")
	return cmdutil.CreateAlias(useAuthToken, "auth use-auth-token")
}

// CheckRepoCmd returns a cobra command that sends a GetPermissions request to
// pachd to determine what permissions a user has on the repo.
func CheckRepoCmd(ctx context.Context, pachCtx *config.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	project := pachCtx.Project
	check := &cobra.Command{
		Use:   "{{alias}} <repo> [<user>]",
		Short: "Check the permissions a user has on 'repo'",
		Long:  "Check the permissions a user has on 'repo'",
		Run: cmdutil.RunBoundedArgs(1, 2, func(args []string) error {
			repoResource := client.NewRepo(project, args[0]).AuthResource()
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var perms *auth.GetPermissionsResponse
			if len(args) == 2 {
				perms, err = c.GetPermissionsForPrincipal(c.Ctx(), &auth.GetPermissionsForPrincipalRequest{
					Resource:  repoResource,
					Principal: args[1],
				})
			} else {
				perms, err = c.GetPermissions(c.Ctx(), &auth.GetPermissionsRequest{
					Resource: repoResource,
				})
			}
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Printf("Roles: %v\nPermissions: %v\n", perms.Roles, perms.Permissions)
			return nil
		}),
	}
	check.Flags().StringVar(&project, "project", project, "The project containing the repo.")
	return cmdutil.CreateAliases(check, "auth check repo", "repos")
}

// SetRepoRoleBindingCmd returns a cobra command that sets the roles for a user on a repo
func SetRepoRoleBindingCmd(ctx context.Context, pachCtx *config.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	project := pachCtx.Project
	setScope := &cobra.Command{
		Use:   "{{alias}} <repo> [role1,role2 | none ] <subject>",
		Short: "Set the roles that 'subject' has on 'repo'",
		Long:  "Set the roles that 'subject' has on 'repo'",
		Run: cmdutil.RunFixedArgs(3, func(args []string) error {
			var roles []string
			if args[1] == "none" {
				roles = []string{}
			} else {
				roles = strings.Split(args[1], ",")
			}

			subject, repo := args[2], args[0]
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			err = c.ModifyRepoRoleBinding(project, repo, subject, roles)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	setScope.Flags().StringVar(&project, "project", project, "The project containing the repo.")
	return cmdutil.CreateAliases(setScope, "auth set repo", "repos")
}

// GetRepoRoleBindingCmd returns a cobra command that gets the role bindings for a repo
func GetRepoRoleBindingCmd(ctx context.Context, pachCtx *config.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	project := pachCtx.Project
	get := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Get the role bindings for 'repo'",
		Long:  "Get the role bindings for 'repo'",
		Run: cmdutil.RunBoundedArgs(1, 1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			repo := args[0]
			resp, err := c.GetRepoRoleBinding(project, repo)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			printRoleBinding(resp)
			return nil
		}),
	}
	get.Flags().StringVar(&project, "project", project, "The project containing the repo.")
	return cmdutil.CreateAliases(get, "auth get repo", "repos")
}

// CheckProjectCmd returns a cobra command that sends a GetPermissions request to
// pachd to determine what permissions a user has on the project.
func CheckProjectCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	check := &cobra.Command{
		Use:   "{{alias}} <project> [user]",
		Short: "Check the permissions a user has on 'project'",
		Long:  "Check the permissions a user has on 'project'",
		Run: cmdutil.RunBoundedArgs(1, 2, func(args []string) error {
			project := client.NewProject(args[0]).AuthResource()
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var perms *auth.GetPermissionsResponse
			if len(args) == 2 {
				perms, err = c.GetPermissionsForPrincipal(c.Ctx(), &auth.GetPermissionsForPrincipalRequest{
					Resource:  project,
					Principal: args[1],
				})
			} else {
				perms, err = c.GetPermissions(c.Ctx(), &auth.GetPermissionsRequest{
					Resource: project,
				})
			}
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Printf("Roles: %v\nPermissions: %v\n", perms.Roles, perms.Permissions)
			return nil
		}),
	}
	return cmdutil.CreateAliases(check, "auth check project")
}

// SetProjectRoleBindingCmd returns a cobra command that sets the roles for a user on a project
func SetProjectRoleBindingCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "{{alias}} <project> [role1,role2 | none ] <subject>",
		Short: "Set the roles that 'subject' has on 'project'",
		Long:  "Set the roles that 'subject' has on 'project'",
		Run: cmdutil.RunFixedArgs(3, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var (
				project *auth.Resource
				user    string
				roles   []string
			)
			project = &auth.Resource{Type: auth.ResourceType_PROJECT, Name: args[0]}
			user = args[2]
			if args[1] != "none" {
				roles = strings.Split(args[1], ",")
			}

			_, err = c.ModifyRoleBinding(c.Ctx(), &auth.ModifyRoleBindingRequest{
				Resource:  project,
				Principal: user,
				Roles:     roles,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAliases(cmd, "auth set project")
}

// GetProjectRoleBindingCmd returns a cobra command that gets the role bindings for a resource
func GetProjectRoleBindingCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	get := &cobra.Command{
		Use:   "{{alias}} <project>",
		Short: "Get the role bindings for 'project'",
		Long:  "Get the role bindings for 'project'",
		Run: cmdutil.RunBoundedArgs(1, 1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			project := args[0]
			resp, err := c.GetProjectRoleBinding(project)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			printRoleBinding(resp)
			return nil
		}),
	}
	return cmdutil.CreateAliases(get, "auth get project")
}

// SetClusterRoleBindingCmd returns a cobra command that sets the roles for a user on a resource
func SetClusterRoleBindingCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	setScope := &cobra.Command{
		Use:   "{{alias}} [role1,role2 | none ] subject",
		Short: "Set the roles that 'subject' has on the 'cluster'",
		Long:  "Set the roles that 'subject' has on the 'cluster'",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			var roles []string
			if args[0] == "none" {
				roles = []string{}
			} else {
				roles = strings.Split(args[0], ",")
			}

			subject := args[1]
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
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
func GetClusterRoleBindingCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	get := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Get the role bindings for 'cluster'",
		Long:  "Get the role bindings for 'cluster'",
		Run: cmdutil.RunBoundedArgs(0, 0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
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

// SetEnterpriseRoleBindingCmd returns a cobra command that sets the roles for a user on a resource
func SetEnterpriseRoleBindingCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	setScope := &cobra.Command{
		Use:   "{{alias}} [role1,role2 | none ] subject",
		Short: "Set the roles that 'subject' has on the enterprise server",
		Long:  "Set the roles that 'subject' has on the enterprise server",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			var roles []string
			if args[0] == "none" {
				roles = []string{}
			} else {
				roles = strings.Split(args[0], ",")
			}

			subject := args[1]
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			err = c.ModifyClusterRoleBinding(subject, roles)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(setScope, "auth set enterprise")
}

// GetEnterpriseRoleBindingCmd returns a cobra command that gets the role bindings for a resource
func GetEnterpriseRoleBindingCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	get := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Get the role bindings for the enterprise server",
		Long:  "Get the role bindings for the enterprise server",
		Run: cmdutil.RunBoundedArgs(0, 0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
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
	return cmdutil.CreateAlias(get, "auth get enterprise")
}

// RotateRootToken returns a cobra command that rotates the auth token for the Root User
func RotateRootToken(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var rootToken string
	rotateRootToken := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Rotate the root user's auth token",
		Long:  "Rotate the root user's auth token",
		Run: cmdutil.RunBoundedArgs(0, 0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			req := &auth.RotateRootTokenRequest{
				RootToken: rootToken,
			}
			resp, err := c.RotateRootToken(c.Ctx(), req)
			if err != nil {
				return err
			}

			fmt.Println("WARNING: DO NOT LOSE THE AUTH TOKEN BELOW. STORE IT SECURELY FOR THE LIFE OF THE CLUSTER." +
				"THIS TOKEN WILL ALWAYS HAVE ADMIN ACCESS TO FIX THE CLUSTER CONFIGURATION.")
			fmt.Printf("Pachyderm auth token:\n%s\n", resp.RootToken)
			return nil
		}),
	}
	rotateRootToken.PersistentFlags().StringVar(&rootToken, "supply-token", "", "An auth token to rotate to. If left blank, one will be auto-generated.")

	return cmdutil.CreateAlias(rotateRootToken, "auth rotate-root-token")
}

// RolesForPermissionCmd lists the roles that would give a user a specific permission
func RolesForPermissionCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	rotateRootToken := &cobra.Command{
		Use:   "{{alias}} <permission>",
		Short: "List roles that grant the given permission",
		Long:  "List roles that grant the given permission",
		Run: cmdutil.RunBoundedArgs(1, 1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			permission, ok := auth.Permission_value[strings.ToUpper(args[0])]
			if !ok {
				return errors.Errorf("unknown permission %q", args[0])
			}

			resp, err := c.GetRolesForPermission(c.Ctx(), &auth.GetRolesForPermissionRequest{Permission: auth.Permission(permission)})
			if err != nil {
				return err
			}

			names := make([]string, len(resp.Roles))
			for i, r := range resp.Roles {
				names[i] = r.Name
			}
			sort.Strings(names)
			fmt.Print(strings.Join(names, "\n"))
			return nil
		}),
	}

	return cmdutil.CreateAlias(rotateRootToken, "auth roles-for-permission")
}

// Cmds returns a list of cobra commands for authenticating and authorizing
// users in an auth-enabled Pachyderm cluster.
func Cmds(mainCtx context.Context, pachCtx *config.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
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
	commands = append(commands, ActivateCmd(mainCtx, pachctlCfg))
	commands = append(commands, DeactivateCmd(mainCtx, pachctlCfg))
	commands = append(commands, LoginCmd(mainCtx, pachctlCfg))
	commands = append(commands, LogoutCmd())
	commands = append(commands, WhoamiCmd(mainCtx, pachctlCfg))
	commands = append(commands, GetRobotTokenCmd(mainCtx, pachctlCfg))
	commands = append(commands, UseAuthTokenCmd())
	commands = append(commands, GetConfigCmd(mainCtx, pachctlCfg))
	commands = append(commands, SetConfigCmd(mainCtx, pachctlCfg))
	commands = append(commands, RevokeCmd(mainCtx, pachctlCfg))
	commands = append(commands, GetGroupsCmd(mainCtx, pachctlCfg))
	commands = append(commands, CheckRepoCmd(mainCtx, pachCtx, pachctlCfg))
	commands = append(commands, GetRepoRoleBindingCmd(mainCtx, pachCtx, pachctlCfg))
	commands = append(commands, SetRepoRoleBindingCmd(mainCtx, pachCtx, pachctlCfg))
	commands = append(commands, CheckProjectCmd(mainCtx, pachctlCfg))
	commands = append(commands, GetProjectRoleBindingCmd(mainCtx, pachctlCfg))
	commands = append(commands, SetProjectRoleBindingCmd(mainCtx, pachctlCfg))
	commands = append(commands, GetClusterRoleBindingCmd(mainCtx, pachctlCfg))
	commands = append(commands, SetClusterRoleBindingCmd(mainCtx, pachctlCfg))
	commands = append(commands, GetEnterpriseRoleBindingCmd(mainCtx, pachctlCfg))
	commands = append(commands, SetEnterpriseRoleBindingCmd(mainCtx, pachctlCfg))
	commands = append(commands, RotateRootToken(mainCtx, pachctlCfg))
	commands = append(commands, RolesForPermissionCmd(mainCtx, pachctlCfg))
	return commands
}
