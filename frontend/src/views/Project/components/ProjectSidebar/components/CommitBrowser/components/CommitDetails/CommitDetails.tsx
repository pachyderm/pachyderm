import {RepoQuery} from '@graphqlTypes';
import {format, fromUnixTime} from 'date-fns';
import React from 'react';
import {useHistory} from 'react-router';

import Description from '@dash-frontend/components/Description';
import useCommit from '@dash-frontend/hooks/useCommit';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {jobRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  LoadingDots,
  Button,
  StatusUpdatedSVG,
  AddCircleSVG,
  CloseCircleSVG,
  Icon,
  ButtonGroup,
} from '@pachyderm/components';

import styles from './CommitDetails.module.css';

type CommitDetailsProps = {
  repo?: RepoQuery['repo'];
  commitId: string;
};

const CommitDetails: React.FC<CommitDetailsProps> = ({commitId, repo}) => {
  const browserHistory = useHistory();
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const {getUpdatedSearchParams} = useUrlQueryState();
  const {branchId, projectId, repoId} = useUrlState();

  const selectedBranch =
    branchId === 'default' ? repo?.branches[0].name : branchId;

  const goToJob = (jobId: string, repoId: string) => {
    const newSearchParams = getUpdatedSearchParams({
      selectedJobs: [jobId],
      pipelineStep: [repoId],
    });

    browserHistory.push(`${jobRoute({projectId}, false)}${newSearchParams}`);
  };

  const {commit, loading} = useCommit({
    args: {
      projectId,
      repoName: repoId,
      branchName: selectedBranch,
      id: commitId,
      withDiff: true,
    },
  });

  if (loading) {
    return (
      <div
        data-testid="CommitDetails__loadingdots"
        className={styles.loadingDots}
      >
        <LoadingDots />
      </div>
    );
  }

  return commit ? (
    <div className={styles.commit}>
      <dl>
        <Description
          loading={loading}
          term="Commit ID"
          data-testid="CommitDetails__id"
        >
          {commit.id}
        </Description>
        {commit.started && commit.finished !== -1 && (
          <Description
            loading={loading}
            term="Start Time"
            data-testid="CommitDetails__start"
          >
            {format(fromUnixTime(commit.started), 'MM/dd/yyyy h:mm:ssaaa')}
          </Description>
        )}
        {commit.finished && commit.finished !== -1 && (
          <Description
            loading={loading}
            term="End Time"
            data-testid="CommitDetails__end"
          >
            {format(fromUnixTime(commit.finished), 'MM/dd/yyyy h:mm:ssaaa')}
          </Description>
        )}
        {commit.description && (
          <Description
            term="Description"
            data-testid="CommitDetails__description"
          >
            {commit.description}
          </Description>
        )}
        {commit.diff &&
          (commit.diff.filesAdded ||
            commit.diff.filesUpdated ||
            commit.diff.filesDeleted) !== 0 && (
            <Description
              term={`File Updates (${commit.diff.size > 0 ? '+' : ''}${
                commit.diff.sizeDisplay
              })`}
              className={styles.diffUpdates}
              data-testid="CommitDetails__fileUpdates"
            >
              {commit.diff.filesAdded > 0 && (
                <div className={styles.filesAdded}>
                  <Icon small color="green" className={styles.commitStatusIcon}>
                    <AddCircleSVG />
                  </Icon>
                  {`${commit.diff.filesAdded} File${
                    commit.diff.filesAdded > 1 ? 's' : ''
                  } added`}
                </div>
              )}
              {commit.diff.filesUpdated > 0 && (
                <div className={styles.filesUpdated}>
                  <Icon small color="green" className={styles.commitStatusIcon}>
                    <StatusUpdatedSVG />
                  </Icon>
                  {`${commit.diff.filesUpdated} File${
                    commit.diff.filesUpdated > 1 ? 's' : ''
                  } updated`}
                </div>
              )}
              {commit.diff.filesDeleted > 0 && (
                <div className={styles.filesDeleted}>
                  <Icon small color="red" className={styles.commitStatusIcon}>
                    <CloseCircleSVG />
                  </Icon>
                  {`${commit.diff.filesDeleted} File${
                    commit.diff.filesDeleted > 1 ? 's' : ''
                  } deleted`}
                </div>
              )}
            </Description>
          )}
      </dl>
      <ButtonGroup>
        {selectedBranch && (
          <Button
            onClick={() =>
              browserHistory.push(
                getPathToFileBrowser({
                  projectId,
                  branchId: selectedBranch,
                  repoId: repoId,
                  commitId: commit.id,
                }),
              )
            }
            disabled={!!repo?.linkedPipeline && commit.finished === -1}
          >
            View Files
          </Button>
        )}
        {commit.hasLinkedJob && (
          <Button
            buttonType="secondary"
            onClick={() => goToJob(commit.id, commit.repoName)}
          >
            Linked Job
          </Button>
        )}
      </ButtonGroup>
    </div>
  ) : (
    <div />
  );
};

export default CommitDetails;
