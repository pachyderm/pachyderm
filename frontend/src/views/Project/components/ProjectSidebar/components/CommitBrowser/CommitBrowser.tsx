import {OriginKind, RepoQuery} from '@graphqlTypes';
import {
  LoadingDots,
  Tooltip,
  PureCheckbox,
  Group,
  Button,
  ButtonGroup,
} from '@pachyderm/components';
import React from 'react';
import {CSSTransition, TransitionGroup} from 'react-transition-group';

import CommitIdCopy from '@dash-frontend/components/CommitIdCopy';
import EmptyState from '@dash-frontend/components/EmptyState';
import {LETS_START_TITLE} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import {COMMITS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import useCommits, {COMMIT_LIMIT} from '@dash-frontend/hooks/useCommits';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {jobRoute} from '@dash-frontend/views/Project/utils/routes';

import styles from './CommitBrowser.module.css';
import BranchBrowser from './components/BranchBrowser';
import CommitTime from './components/CommitTime';

const emptyRepoMessage = 'Commit your first file on this repo!';

type CommitBrowserProps = {
  repo?: RepoQuery['repo'];
  repoBaseRef: React.RefObject<HTMLDivElement>;
};

const CommitBrowser: React.FC<CommitBrowserProps> = ({repo, repoBaseRef}) => {
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const {branchId, projectId, repoId} = useUrlState();
  const [hideAutoCommits, handleHideAutoCommitChange] = useLocalProjectSettings(
    {projectId, key: 'hide_auto_commits'},
  );

  const {commits, loading} = useCommits({
    args: {
      projectId,
      repoName: repoId,
      pipelineName: repo?.linkedPipeline?.name,
      originKind: hideAutoCommits ? OriginKind.USER : undefined,
      branchName: branchId,
      number: COMMIT_LIMIT,
    },
    skip: !repo || repo.branches.length === 0,
  });

  if (loading) {
    return (
      <div
        data-testid="CommitBrowser__loadingdots"
        className={styles.loadingDots}
      >
        <LoadingDots />
      </div>
    );
  }

  if (
    !hideAutoCommits &&
    (repo?.branches?.length || 0) <= 1 &&
    !commits?.length
  ) {
    return <EmptyState title={LETS_START_TITLE} message={emptyRepoMessage} />;
  }

  return (
    <>
      <BranchBrowser repo={repo} repoBaseRef={repoBaseRef} />
      <div className={styles.autoCommits}>
        <PureCheckbox
          selected={!hideAutoCommits}
          small
          label="Auto Commits"
          onChange={() => handleHideAutoCommitChange(!hideAutoCommits)}
        />
      </div>
      <div className={styles.commits}>
        {commits?.length ? (
          <TransitionGroup>
            {commits.map((commit) => {
              // only animate in if commit is younger than poll interval
              const freshCommit =
                new Date().getTime() -
                  (commit.started !== -1 ? commit.started : 0) * 1000 <
                COMMITS_POLL_INTERVAL_MS;
              return (
                <CSSTransition
                  timeout={400}
                  classNames={
                    freshCommit
                      ? {
                          enterActive: styles.commitEnterActive,
                          enterDone: styles.commitEnterDone,
                          exitActive: styles.commitExit,
                          exitDone: styles.commitExitActive,
                        }
                      : {}
                  }
                  key={commit.id}
                >
                  <div
                    className={styles.commit}
                    data-testid="CommitBrowser__commit"
                  >
                    <strong>
                      <CommitTime commit={commit} />
                      {`, ${commit.sizeDisplay}`}
                    </strong>
                    <dl className={styles.commitInfo}>
                      {commit.description ? (
                        <Tooltip
                          tooltipText={commit.description}
                          placement="left"
                          tooltipKey="description"
                        >
                          <Group
                            spacing={8}
                            className={styles.commitData}
                            align="center"
                          >
                            <CommitIdCopy
                              small
                              longId
                              commit={commit.id}
                              clickable
                            />
                          </Group>
                        </Tooltip>
                      ) : (
                        <Group spacing={8} className={styles.commitData}>
                          <CommitIdCopy
                            small
                            longId
                            commit={commit.id}
                            clickable
                          />
                        </Group>
                      )}
                      <dt>
                        <ButtonGroup>
                          <Button
                            buttonType="secondary"
                            to={getPathToFileBrowser({
                              projectId,
                              branchId,
                              repoId: repoId,
                              commitId: commit.id,
                            })}
                          >
                            View Files
                          </Button>
                          {commit.hasLinkedJob && (
                            <Button
                              buttonType="ghost"
                              to={jobRoute({
                                projectId,
                                jobId: commit.id,
                                pipelineId: repo?.name,
                              })}
                            >
                              Linked Job
                            </Button>
                          )}
                        </ButtonGroup>
                      </dt>
                    </dl>
                  </div>
                </CSSTransition>
              );
            })}
          </TransitionGroup>
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
