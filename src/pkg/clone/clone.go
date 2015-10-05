package clone //import "go.pachyderm.com/pachyderm/src/pkg/clone"

import (
	"fmt"

	"gopkg.in/libgit2/git2go.v22"
)

func GithubClone(
	dirPath string,
	user string,
	repository string,
	branch string,
	commitID string,
	accessToken string,
) error {
	gitRepository, err := git.Clone(
		fmt.Sprintf("https://github.com/%s/%s", user, repository),
		dirPath,
		&git.CloneOptions{
			RemoteCallbacks: getRemoteCallbacks(user, accessToken),
			CheckoutBranch:  branch,
		},
	)
	if err != nil {
		return err
	}
	return checkoutCommitID(gitRepository, commitID)
}

func getRemoteCallbacks(user string, accessToken string) *git.RemoteCallbacks {
	if accessToken == "" {
		return nil
	}
	return &git.RemoteCallbacks{
		CredentialsCallback: func(url string, username_from_url string, allowed_types git.CredType) (git.ErrorCode, *git.Cred) {
			_, cred := git.NewCredUserpassPlaintext(user, accessToken)
			return git.ErrOk, &cred
		},
	}
}

func checkoutCommitID(gitRepository *git.Repository, commitID string) error {
	if commitID == "" {
		return nil
	}
	oid, err := git.NewOid(commitID)
	if err != nil {
		return err
	}
	if err := gitRepository.SetHeadDetached(
		oid,
		nil,
		"",
	); err != nil {
		return err
	}
	return gitRepository.CheckoutHead(
		&git.CheckoutOpts{
			Strategy: git.CheckoutForce,
		},
	)
}
