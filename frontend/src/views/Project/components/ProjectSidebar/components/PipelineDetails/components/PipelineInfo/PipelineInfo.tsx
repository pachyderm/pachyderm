import capitalize from 'lodash/capitalize';
import React from 'react';

import Description from '@dash-frontend/components/Description';
import PipelineStateComponent from '@dash-frontend/components/PipelineState';
import {RepoLink} from '@dash-frontend/components/ResourceLink';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import extractAndShortenIds from '@dash-frontend/lib/extractAndShortenIds';

import styles from './PipelineInfo.module.css';

const PipelineInfo: React.FC = () => {
  const {pipelineId} = useUrlState();
  const {searchParams} = useUrlQueryState();
  const {pipeline, loading, isServiceOrSpout} = useCurrentPipeline();

  const globalId = searchParams.globalIdFilter;

  return (
    <dl className={styles.base}>
      {!globalId && (
        <Description term="" loading={loading}>
          {pipeline && <PipelineStateComponent state={pipeline.state} />}
        </Description>
      )}

      <Description term="Output Repo" loading={loading}>
        <div className={styles.repoGroup}>
          <RepoLink name={pipelineId} />
        </div>
      </Description>

      {!globalId && pipeline?.reason && (
        <Description term="Failure Reason" loading={loading}>
          {extractAndShortenIds(pipeline.reason)}
        </Description>
      )}

      <Description term="Pipeline Type" loading={loading}>
        {capitalize(pipeline?.type)}
      </Description>

      {!isServiceOrSpout && (
        <>
          <Description term="Datum Timeout" loading={loading}>
            {pipeline?.datumTimeoutS
              ? `${pipeline.datumTimeoutS} seconds`
              : 'N/A'}
          </Description>

          <Description term="Datum Tries" loading={loading}>
            {pipeline?.datumTries}
          </Description>

          <Description term="Job Timeout" loading={loading}>
            {pipeline?.jobTimeoutS ? `${pipeline.jobTimeoutS} seconds` : 'N/A'}
          </Description>
        </>
      )}
      <Description term="Output Branch" loading={loading}>
        {pipeline?.outputBranch}
      </Description>

      {!isServiceOrSpout && (
        <>
          <Description term="Egress" loading={loading}>
            {pipeline?.egress ? 'Yes' : 'No'}
          </Description>

          <Description term="S3 Output Repo" loading={loading}>
            {pipeline?.s3OutputRepo || 'N/A'}
          </Description>
        </>
      )}
    </dl>
  );
};

export default PipelineInfo;
