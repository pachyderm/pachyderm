import useCommit from '@dash-frontend/hooks/useCommit';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useRepoDetails = () => {
  const {repoId, projectId, commitId} = useUrlState();
  const {loading: repoLoading, repo, error: repoError} = useCurrentRepo();

  const {
    commit,
    loading: commitLoading,
    error: commitError,
  } = useCommit({
    args: {
      projectId,
      repoName: repoId,
      id: commitId ? commitId : '',
      withDiff: true,
    },
  });

  const currentRepoLoading =
    repoLoading || commitLoading || repoId !== repo?.id;

  return {
    repo,
    commit,
    currentRepoLoading,
    commitError,
    repoError,
  };
};

export default useRepoDetails;
