import {SkeletonDisplayText, Link} from '@pachyderm/components';
import {format, fromUnixTime} from 'date-fns';
import React from 'react';

import Description from '@dash-frontend/components/Description';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

import Title from '../Title';

import styles from './RepoDetails.module.css';

const RepoDetails = () => {
  const {projectId, dagId} = useUrlState();
  const {loading, repo} = useCurrentRepo();

  return (
    <div className={styles.base}>
      {loading ? (
        <SkeletonDisplayText data-testid={'RepoDetails__RepoNameSkeleton'} />
      ) : (
        <Title>{repo?.name}</Title>
      )}

      <dl>
        <Description loading={loading} term="Created">
          {repo ? format(fromUnixTime(repo.createdAt), 'MM/d/yyyy') : 'N/A'}
        </Description>
        <Description loading={loading} term="Linked Pipeline">
          {repo?.linkedPipeline ? (
            <Link
              to={pipelineRoute({
                projectId,
                dagId,
                pipelineId: repo?.linkedPipeline?.id,
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
    </div>
  );
};

export default RepoDetails;
