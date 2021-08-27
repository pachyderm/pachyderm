import {SkeletonDisplayText, Link} from '@pachyderm/components';
import {format, fromUnixTime} from 'date-fns';
import React, {useRef} from 'react';

import Description from '@dash-frontend/components/Description';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

import CommitBrowser from '../CommitBrowser';
import Title from '../Title';

import styles from './RepoDetails.module.css';

const RepoDetails = () => {
  const {projectId} = useUrlState();
  const {loading, repo} = useCurrentRepo();
  const repoBaseRef = useRef<HTMLDivElement>(null);

  return (
    <div className={styles.base} ref={repoBaseRef}>
      <div className={styles.title}>
        {loading ? (
          <SkeletonDisplayText data-testid={'RepoDetails__RepoNameSkeleton'} />
        ) : (
          <Title>{repo?.name}</Title>
        )}
      </div>

      <dl className={styles.repoInfo}>
        <Description loading={loading} term="Created">
          {repo ? format(fromUnixTime(repo.createdAt), 'MM/d/yyyy') : 'N/A'}
        </Description>
        <Description loading={loading} term="Linked Pipeline">
          {repo?.linkedPipeline ? (
            <Link
              to={pipelineRoute({
                projectId,
                pipelineId: repo?.linkedPipeline?.id,
                tabId: 'info',
              })}
            >
              {repo.linkedPipeline?.name}
            </Link>
          ) : (
            'N/A'
          )}
        </Description>
        <Description loading={loading} term="Description">
          {repo?.description ? repo.description : 'N/A'}
        </Description>
        <Description term="Size">
          {!loading ? repo?.sizeDisplay : <SkeletonDisplayText />}
        </Description>
      </dl>
      <div className={styles.divider} />
      <CommitBrowser repo={repo} repoBaseRef={repoBaseRef} loading={loading} />
    </div>
  );
};

export default RepoDetails;
