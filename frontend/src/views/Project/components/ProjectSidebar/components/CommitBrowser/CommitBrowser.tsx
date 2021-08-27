import {RepoQuery} from '@graphqlTypes';
import {Link, LoadingDots} from '@pachyderm/components';
import formatDistanceToNow from 'date-fns/formatDistanceToNow';
import React, {useMemo} from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import {LETS_START_TITLE} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  fileBrowserRoute,
  jobRoute,
} from '@dash-frontend/views/Project/utils/routes';

import styles from './CommitBrowser.module.css';
import BranchBrowser from './components/BranchBrowser';

const emptyRepoMessage = 'Commit your first file on this repo!';

type CommitBrowserProps = {
  repo?: RepoQuery['repo'];
  repoBaseRef: React.RefObject<HTMLDivElement>;
  loading: boolean;
};

const CommitBrowser: React.FC<CommitBrowserProps> = ({
  loading,
  repo,
  repoBaseRef,
}) => {
  const {branchId, projectId, repoId} = useUrlState();
  const currentCommits = useMemo(
    () =>
      (repo?.commits || []).filter(
        (commit) => commit?.branch?.name === branchId,
      ),
    [branchId, repo],
  );

  if (loading) {
    return (
      <div data-testid="CommitBrowser__loadingdots">
        <LoadingDots />
      </div>
    );
  }

  if (repo?.commits.length === 0) {
    return <EmptyState title={LETS_START_TITLE} message={emptyRepoMessage} />;
  }

  return (
    <>
      <BranchBrowser repo={repo} repoBaseRef={repoBaseRef} />
      <div className={styles.commits}>
        {currentCommits.length ? (
          currentCommits.map((commit) => {
            return (
              <div className={styles.commit} key={commit.id}>
                <div className={styles.commitTime}>
                  {commit.finished
                    ? `Committed
                  ${formatDistanceToNow(commit.finished * 1000, {
                    addSuffix: true,
                  })}`
                    : `Commit started ${formatDistanceToNow(
                        commit.started * 1000,
                        {
                          addSuffix: true,
                        },
                      )}`}
                </div>
                <dl className={styles.commitInfo}>
                  <dt className={styles.commitData}>ID {commit.id}</dt>
                  <dt className={styles.commitData}>
                    Size {commit.sizeDisplay}
                  </dt>
                  <dt className={styles.commitData}>
                    {commit.hasLinkedJob && (
                      <Link
                        to={jobRoute({
                          projectId,
                          jobId: commit.id,
                          pipelineId: repo?.name,
                        })}
                      >
                        Linked Job
                      </Link>
                    )}
                  </dt>
                  <dt className={styles.commitData}>
                    <Link
                      to={fileBrowserRoute({
                        projectId,
                        branchId,
                        repoId: repoId,
                        commitId: commit.id,
                      })}
                    >
                      View Files
                    </Link>
                  </dt>
                </dl>
              </div>
            );
          })
        ) : (
          <div className={styles.empty}>
            There are no commits for this branch
          </div>
        )}
      </div>
    </>
  );
};

export default CommitBrowser;
