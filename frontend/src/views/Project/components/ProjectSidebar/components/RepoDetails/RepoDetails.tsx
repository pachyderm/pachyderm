import {format, fromUnixTime} from 'date-fns';
import React from 'react';
import {Helmet} from 'react-helmet';

import CommitIdCopy from '@dash-frontend/components/CommitIdCopy';
import Description from '@dash-frontend/components/Description';
import {standardFormat} from '@dash-frontend/constants/dateFormats';
import {
  SkeletonDisplayText,
  ElephantEmptyState,
  Link,
  CaptionTextSmall,
} from '@pachyderm/components';

import Title from '../Title';

import CommitDetails from './components/CommitDetails';
import CommitList from './components/CommitList';
import useRepoDetails from './hooks/useRepoDetails';
import styles from './RepoDetails.module.css';

const RepoDetails: React.FC = () => {
  const {repo, commit, repoError, currentRepoLoading} = useRepoDetails();

  if (!currentRepoLoading && repoError) {
    return (
      <div className={styles.emptyRepoMessage}>
        <ElephantEmptyState className={styles.emptyElephant} />
        <h5>Unable to load Repo and commit data</h5>
        <p>
          We weren&apos;t able to fetch any information about this repo,
          including its latest commit. Please try refreshing this page. If this
          issue keeps happening, contact our customer team.
        </p>
      </div>
    );
  }

  return (
    <div className={styles.base}>
      <Helmet>
        <title>Repo - Pachyderm Console</title>
      </Helmet>
      <div className={styles.titleSection}>
        {currentRepoLoading ? (
          <SkeletonDisplayText data-testid="RepoDetails__repoNameSkeleton" />
        ) : (
          <Title>{repo?.name}</Title>
        )}
        {repo?.description && (
          <div className={styles.description}>{repo?.description}</div>
        )}
        <Description
          loading={currentRepoLoading}
          term="Repo Created"
          error={repoError}
        >
          {repo ? format(fromUnixTime(repo.createdAt), standardFormat) : 'N/A'}
        </Description>
        {(currentRepoLoading || commit) && (
          <Description
            loading={currentRepoLoading}
            term="Most Recent Commit Start"
          >
            {commit
              ? format(fromUnixTime(commit.started), standardFormat)
              : 'N/A'}
          </Description>
        )}
        {(currentRepoLoading || commit) && (
          <Description
            loading={currentRepoLoading}
            term="Most Recent Commit ID"
          >
            {commit ? (
              <CommitIdCopy commit={commit.id} clickable small longId />
            ) : (
              'N/A'
            )}
          </Description>
        )}
      </div>

      {!currentRepoLoading && (!commit || repo?.branches.length === 0) && (
        <div className={styles.emptyRepoMessage}>
          <ElephantEmptyState className={styles.emptyElephant} />
          <Title>This repo doesn&apos;t have any branches</Title>
          <p>
            This is normal for new repositories, but we still wanted to notify
            you because Pachyderm didn&apos;t detect a branch on our end.{' '}
            <Link
              externalLink
              to="https://docs.pachyderm.com/2.3.x/concepts/data-concepts/branch/"
            >
              Try creating a branch and pushing a commit.
            </Link>
          </p>
        </div>
      )}

      {commit?.id && (
        <>
          <CaptionTextSmall className={styles.commitDetailsLabel}>
            Current Commit Stats
          </CaptionTextSmall>
          <CommitDetails commit={commit} repo={repo} />
        </>
      )}
      {repo && commit && <CommitList repo={repo} />}
    </div>
  );
};

export default RepoDetails;
