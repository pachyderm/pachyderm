import {FileDocSVG, Link} from '@pachyderm/components';
import {
  fromUnixTime,
  formatDistanceToNow,
  formatDistanceStrict,
} from 'date-fns';
import React, {useMemo} from 'react';
import {useRouteMatch, Redirect} from 'react-router';

import {NOT_FOUND_ERROR_CODE} from '@dash-backend/lib/types';
import Description from '@dash-frontend/components/Description';
import JSONBlock from '@dash-frontend/components/JSONBlock';
import PipelineInput from '@dash-frontend/components/PipelineInput';
import {useJob} from '@dash-frontend/hooks/useJob';
import readableJobState from '@dash-frontend/lib/readableJobState';
import {ProjectRouteParams} from '@dash-frontend/lib/types';
import {
  jobRoute,
  logsViewerJobRoute,
} from '@dash-frontend/views/Project/utils/routes';

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
    return job?.startedAt
      ? formatDistanceToNow(fromUnixTime(job.startedAt), {
          addSuffix: true,
        })
      : 'N/A';
  }, [job?.startedAt]);

  const duration = useMemo(() => {
    return job?.finishedAt && job?.startedAt
      ? formatDistanceStrict(
          fromUnixTime(job.finishedAt),
          fromUnixTime(job.startedAt),
        )
      : 'N/A';
  }, [job?.startedAt, job?.finishedAt]);

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
      <div className={styles.readLogsWrapper}>
        <Link
          small
          to={logsViewerJobRoute({
            projectId,
            jobId: jobId,
            pipelineId: pipelineId,
          })}
        >
          Read Logs{' '}
          <FileDocSVG className={styles.readLogsSvg} width={20} height={24} />
        </Link>
      </div>

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

      <hr className={styles.divider} />

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
    </dl>
  );
};

export default InfoPanel;
