import {fromUnixTime, formatDistanceToNow, formatDistance} from 'date-fns';
import React, {useMemo} from 'react';
import {useRouteMatch, Redirect} from 'react-router';

import {NOT_FOUND_ERROR_CODE} from '@dash-backend/lib/types';
import Description from '@dash-frontend/components/Description';
import JSONBlock from '@dash-frontend/components/JSONBlock';
import PipelineInput from '@dash-frontend/components/PipelineInput';
import {useJob} from '@dash-frontend/hooks/useJob';
import readableJobState from '@dash-frontend/lib/readableJobState';
import {ProjectRouteParams} from '@dash-frontend/lib/types';
import {jobRoute} from '@dash-frontend/views/Project/utils/routes';

import styles from './InfoPanel.module.css';

interface InfoPanelParams extends ProjectRouteParams {
  jobId: string;
  pipelineId: string;
}

const InfoPanel = () => {
  const {
    params: {jobId, projectId, pipelineId},
  } = useRouteMatch<InfoPanelParams>();

  const {job, loading, error} = useJob({
    id: jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const transformString = useMemo(() => {
    if (job?.transform) {
      return (
        <JSONBlock>
          {JSON.stringify(
            {
              image: job.transform.image,
              cmd: job.transform.cmdList,
            },
            null,
            2,
          )}
        </JSONBlock>
      );
    }
    return 'N/A';
  }, [job?.transform]);

  const started = useMemo(() => {
    return job?.createdAt
      ? formatDistanceToNow(fromUnixTime(job.createdAt), {
          addSuffix: true,
        })
      : 'N/A';
  }, [job?.createdAt]);

  const duration = useMemo(() => {
    return job?.finishedAt && job?.createdAt
      ? formatDistance(
          fromUnixTime(job.finishedAt),
          fromUnixTime(job.createdAt),
        )
      : 'N/A';
  }, [job?.createdAt, job?.finishedAt]);

  if (
    error?.graphQLErrors.some(
      (err) => err.extensions?.code === NOT_FOUND_ERROR_CODE,
    )
  ) {
    return (
      <Redirect
        to={jobRoute({
          projectId,
          jobId,
        })}
      />
    );
  }

  return (
    <dl className={styles.base}>
      <Description
        term="Inputs"
        lines={9}
        loading={loading}
        data-testid="InfoPanel__inputs"
      >
        {job?.inputString ? (
          <PipelineInput
            inputString={job.inputString}
            branchId={job.inputBranch || 'master'}
            projectId={projectId}
          />
        ) : (
          'N/A'
        )}
      </Description>

      <Description
        term="Transform"
        loading={loading}
        lines={3}
        data-testid="InfoPanel__transform"
      >
        {transformString}
      </Description>

      <hr className={styles.divider} />

      <Description term="ID" loading={loading} data-testid="InfoPanel__id">
        {job?.id}
      </Description>
      <Description
        term="Pipeline"
        loading={loading}
        data-testid="InfoPanel__pipeline"
      >
        {job?.pipelineName}
      </Description>
      <Description
        term="State"
        loading={loading}
        data-testid="InfoPanel__state"
      >
        {job?.state ? readableJobState(job.state) : 'N/A'}
      </Description>
      <Description
        term="Started"
        loading={loading}
        data-testid="InfoPanel__started"
      >
        {started}
      </Description>
      <Description
        term="Duration"
        loading={loading}
        data-testid="InfoPanel__duration"
      >
        {duration}
      </Description>
    </dl>
  );
};

export default InfoPanel;
