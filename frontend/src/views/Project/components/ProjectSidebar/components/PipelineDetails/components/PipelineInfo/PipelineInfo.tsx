import capitalize from 'lodash/capitalize';
import parse from 'parse-duration';
import React from 'react';

import Description from '@dash-frontend/components/Description';
import PipelineStateComponent from '@dash-frontend/components/PipelineState';
import {RepoLink} from '@dash-frontend/components/ResourceLink';
import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import extractAndShortenIds from '@dash-frontend/lib/extractAndShortenIds';
import readablePipelineType from '@dash-frontend/lib/readablePipelineType';
import {Link} from '@pachyderm/components';

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
        {capitalize(readablePipelineType(pipeline?.type))}
      </Description>

      {!isServiceOrSpout && (
        <>
          <Description term="Datum Timeout" loading={loading}>
            {pipeline?.details?.datumTimeout
              ? `${parse(pipeline?.details?.datumTimeout, 's')} seconds`
              : 'N/A'}
          </Description>

          <Description term="Datum Tries" loading={loading}>
            {pipeline?.details?.datumTries}
          </Description>

          <Description term="Job Timeout" loading={loading}>
            {pipeline?.details?.jobTimeout
              ? `${parse(pipeline.details?.jobTimeout, 's')} seconds`
              : 'N/A'}
          </Description>
        </>
      )}
      <Description term="Output Branch" loading={loading}>
        {pipeline?.details?.outputBranch}
      </Description>

      {pipeline?.details?.service?.type === 'LoadBalancer' && (
        <Description term="Service IP" loading={loading}>
          <Link externalLink to={pipeline?.details?.service?.ip}>
            {pipeline?.details?.service?.ip}
          </Link>
        </Description>
      )}

      {!isServiceOrSpout && (
        <>
          <Description term="Egress" loading={loading}>
            {pipeline?.details?.egress ? 'Yes' : 'No'}
          </Description>

          <Description term="S3 Output Repo" loading={loading}>
            {(pipeline?.details?.s3Out &&
              pipeline?.pipeline &&
              `s3://${pipeline?.pipeline.name}`) ||
              'N/A'}
          </Description>
        </>
      )}
    </dl>
  );
};

export default PipelineInfo;
