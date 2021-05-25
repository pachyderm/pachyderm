import {Link} from '@pachyderm/components';
import formatDistanceToNow from 'date-fns/formatDistanceToNow';
import React, {useMemo} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {RepoQuery} from '@graphqlTypes';

import styles from './CommitBrowser.module.css';
import BranchBrowser from './components/BranchBrowser';

type CommitBrowserProps = {
  repo?: RepoQuery['repo'];
  repoBaseRef: React.RefObject<HTMLDivElement>;
};

const CommitBrowser: React.FC<CommitBrowserProps> = ({repo, repoBaseRef}) => {
  const {branchId, projectId, repoId} = useUrlState();
  const currentCommits = useMemo(
    () =>
      (repo?.commits || []).filter(
        (commit) => commit?.branch?.name === branchId,
      ),
    [branchId, repo],
  );

  return (
    <>
      <BranchBrowser repo={repo} repoBaseRef={repoBaseRef} />
      <div className={styles.commits}>
        {currentCommits.length ? (
          currentCommits.map((commit) => {
            return (
              <div className={styles.commit} key={commit.id}>
                <div className={styles.commitTime}>
                  Committed{' '}
                  {formatDistanceToNow(commit.finished * 1000, {
                    addSuffix: true,
                  })}
                </div>
                <dl className={styles.commitInfo}>
                  <dt className={styles.commitData}>ID {commit.id}</dt>
                  <dt className={styles.commitData}>
                    Size {commit.sizeDisplay}
                  </dt>
                  <dt className={styles.commitData}>
                    <a href="#TODO">Linked Job</a>
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
