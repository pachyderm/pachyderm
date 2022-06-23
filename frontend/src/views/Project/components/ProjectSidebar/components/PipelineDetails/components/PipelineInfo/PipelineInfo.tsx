import capitalize from 'lodash/capitalize';
import React from 'react';

import Description from '@dash-frontend/components/Description';
import PipelineStateComponent from '@dash-frontend/components/PipelineState';
import {RepoLink} from '@dash-frontend/components/ResourceLink';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import extractAndShortenIds from '@dash-frontend/lib/extractAndShortenIds';

import styles from './PipelineInfo.module.css';

type PipelineDetailsProps = {
  dagLinks?: Record<string, string[]>;
};

const PipelineInfo: React.FC<PipelineDetailsProps> = ({dagLinks}) => {
  const {pipelineId} = useUrlState();
  const {pipeline, loading} = useCurrentPipeline();

  return (
    <dl className={styles.base}>
      {dagLinks && dagLinks[pipeline?.name || ''] && (
        <Description loading={loading} term="Inputs">
          <div className={styles.repoGroup}>
            {dagLinks[pipeline?.name || ''].map((name) => (
              <RepoLink name={name} key={name} />
            ))}
          </div>
        </Description>
      )}

      <Description term="Output Repo" loading={loading}>
        <div className={styles.repoGroup}>
          <RepoLink name={pipelineId} />
        </div>
      </Description>

      <Description term="Pipeline Status" loading={loading}>
        {pipeline && <PipelineStateComponent state={pipeline.state} />}
      </Description>

      {pipeline?.reason && (
        <Description term="Status Reason" loading={loading}>
          {extractAndShortenIds(pipeline.reason)}
        </Description>
      )}

      <Description term="Pipeline Type" loading={loading}>
        {capitalize(pipeline?.type)}
      </Description>

      <Description term="Description" loading={loading}>
        {pipeline?.description ? pipeline.description : 'N/A'}
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
