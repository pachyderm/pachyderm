import {Link} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import React from 'react';

import Description from '@dash-frontend/components/Description';
import PipelineStateComponent from '@dash-frontend/components/PipelineState';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';

import styles from './PipelineInfo.module.css';

const PipelineInfo = () => {
  const {projectId, pipelineId} = useUrlState();
  const {pipeline, loading} = useCurrentPipeline();

  return (
    <dl className={styles.base}>
      <Description term="Pipeline Status" loading={loading}>
        {pipeline && <PipelineStateComponent state={pipeline.state} />}
      </Description>

      <Description term="Pipeline Type" loading={loading}>
        {capitalize(pipeline?.type)}
      </Description>

      <Description term="Description" loading={loading}>
        {pipeline?.description ? pipeline.description : 'N/A'}
      </Description>

      <Description term="Output Repo" loading={loading}>
        <Link
          to={repoRoute({
            branchId: 'master',
            projectId,
            repoId: pipelineId,
          })}
        >
          {pipeline?.name}
        </Link>
      </Description>

      <Description term="Cache Size" loading={loading}>
        {pipeline?.cacheSize}
      </Description>

      <Description term="Datum Timeout" loading={loading}>
        {pipeline?.datumTimeoutS ? `${pipeline.datumTimeoutS} seconds` : 'N/A'}
      </Description>

      <Description term="Datum Tries" loading={loading}>
        {pipeline?.datumTries}
      </Description>

      <Description term="Job Timeout" loading={loading}>
        {pipeline?.jobTimeoutS ? `${pipeline.jobTimeoutS} seconds` : 'N/A'}
      </Description>

      <Description term="Output Branch" loading={loading}>
        {pipeline?.outputBranch}
      </Description>

      <Description term="Egress" loading={loading}>
        {pipeline?.egress ? 'Yes' : 'No'}
      </Description>

      <Description term="S3 Output Repo" loading={loading}>
        {pipeline?.s3OutputRepo || 'N/A'}
      </Description>
    </dl>
  );
};

export default PipelineInfo;
