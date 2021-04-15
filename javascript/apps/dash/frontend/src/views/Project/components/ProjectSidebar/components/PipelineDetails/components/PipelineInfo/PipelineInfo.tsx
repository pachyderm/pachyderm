import {SkeletonDisplayText, SkeletonBodyText} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import React from 'react';
import {Link} from 'react-router-dom';

import PipelineState from '@dash-frontend/components/PipelineState';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';

import Description from '../Description';

import styles from './PipelineInfo.module.css';

const Skeleton = () => {
  return (
    <dl>
      <Description term="Pipeline Status">
        <SkeletonDisplayText className={styles.skeleton} />
      </Description>
      <Description term="Pipeline Type">
        <SkeletonDisplayText className={styles.skeleton} />
      </Description>
      <Description term="Description">
        <SkeletonBodyText lines={2} />
      </Description>
      <Description term="Output Repo">
        <SkeletonDisplayText className={styles.skeleton} />
      </Description>
      <Description term="Cache Size">
        <SkeletonDisplayText className={styles.skeleton} />
      </Description>
      <Description term="Datum Timeout">
        <SkeletonDisplayText className={styles.skeleton} />
      </Description>
      <Description term="Datum Tries">
        <SkeletonDisplayText className={styles.skeleton} />
      </Description>
      <Description term="Job Timeout">
        <SkeletonDisplayText className={styles.skeleton} />
      </Description>
      <Description term="Enable Stats">
        <SkeletonDisplayText className={styles.skeleton} />
      </Description>
      <Description term="Output Branch">
        <SkeletonDisplayText className={styles.skeleton} />
      </Description>
    </dl>
  );
};

const PipelineInfo = () => {
  const {projectId} = useUrlState();
  const {pipeline, loading} = useCurrentPipeline();

  if (loading || !pipeline) {
    return <Skeleton />;
  }

  return (
    <dl>
      <Description term="Pipeline Status">
        <PipelineState state={pipeline.state} />
      </Description>

      <Description term="Pipeline Type">
        {capitalize(pipeline.type)}
      </Description>

      <Description term="Description">
        {pipeline.description ? pipeline.description : 'N/A'}
      </Description>

      <Description term="Output Repo">
        <Link to={repoRoute({projectId, repoId: pipeline.name})}>
          {pipeline.name} output repo
        </Link>
      </Description>

      <Description term="Cache Size">{pipeline.cacheSize}</Description>

      <Description term="Datum Timeout">
        {pipeline.datumTimeoutS ? `${pipeline.datumTimeoutS} seconds` : 'N/A'}
      </Description>

      <Description term="Datum Tries">{pipeline.datumTries}</Description>

      <Description term="Job Timeout">
        {pipeline.jobTimeoutS ? `${pipeline.jobTimeoutS} seconds` : 'N/A'}
      </Description>

      <Description term="Enable Stats">
        {pipeline.enableStats ? 'Yes' : 'No'}
      </Description>

      <Description term="Output Branch">{pipeline.outputBranch}</Description>
    </dl>
  );
};

export default PipelineInfo;
