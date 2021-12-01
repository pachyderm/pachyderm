import {NOT_FOUND_ERROR_CODE} from '@dash-backend/lib/types';
import {ButtonLink, DocumentSVG, Link, Icon} from '@pachyderm/components';
import {
  fromUnixTime,
  formatDistanceToNow,
  formatDistanceStrict,
} from 'date-fns';
import React, {useCallback, useMemo} from 'react';
import {useRouteMatch, Redirect, useLocation} from 'react-router';

import Description from '@dash-frontend/components/Description';
import JSONBlock from '@dash-frontend/components/JSONBlock';
import JsonSpec from '@dash-frontend/components/JsonSpec';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import extractAndShortenIds from '@dash-frontend/lib/extractAndShortenIds';
import readableJobState from '@dash-frontend/lib/readableJobState';
import {ProjectRouteParams} from '@dash-frontend/lib/types';
import {
  fileBrowserRoute,
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
  const {pathname} = useLocation();
  const {setUrlFromViewState} = useUrlQueryState();

  const {job, loading, error} = useJob({
    id: jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const handleCommitClick = useCallback(() => {
    if (job?.outputBranch) {
      setUrlFromViewState(
        {prevFileBrowserPath: pathname},
        fileBrowserRoute(
          {
            projectId,
            commitId: job.id,
            branchId: job.outputBranch,
            repoId: job.pipelineName,
          },
          false,
        ),
      );
    }
  }, [setUrlFromViewState, job, pathname, projectId]);

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
          <span className={styles.readLogsText}>
            Read Logs{' '}
            <Icon small color="plum">
              <DocumentSVG />
            </Icon>
          </span>
        </Link>
      </div>

      <Description
        term="Output Commit"
        loading={loading}
        data-testid="InfoPanel__id"
      >
        {job?.outputBranch ? (
          <ButtonLink
            onClick={handleCommitClick}
            data-testid="InfoPanel__commitLink"
          >
            {job?.id}
          </ButtonLink>
        ) : (
          <>{job?.id}</>
        )}
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

      {job?.reason && (
        <Description
          term="Reason"
          loading={loading}
          data-testid="InfoPanel__reason"
        >
          {extractAndShortenIds(job.reason)}
        </Description>
      )}

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

      <Description
        term="Datums Processed"
        loading={loading}
        data-testid="InfoPanel__processed"
      >
        {job?.dataProcessed}
      </Description>

      <Description
        term="Datums Skipped"
        loading={loading}
        data-testid="InfoPanel__skipped"
      >
        {job?.dataSkipped}
      </Description>

      <Description
        term="Datums Failed"
        loading={loading}
        data-testid="InfoPanel__failed"
      >
        {job?.dataFailed}
      </Description>

      <Description
        term="Datums Recovered"
        loading={loading}
        data-testid="InfoPanel__recovered"
      >
        {job?.dataRecovered}
      </Description>

      <Description
        term="Datums Total"
        loading={loading}
        data-testid="InfoPanel__total"
      >
        {job?.dataTotal}
      </Description>

      <hr className={styles.divider} />

      <Description
        term="Inputs"
        lines={9}
        loading={loading}
        data-testid="InfoPanel__inputs"
      >
        {job?.inputString ? (
          <JsonSpec
            jsonString={job.inputString}
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

      <Description
        term="Details"
        lines={9}
        loading={loading}
        data-testid="InfoPanel__details"
      >
        {job?.jsonDetails ? (
          <JsonSpec jsonString={job.jsonDetails} projectId={projectId} />
        ) : (
          'N/A'
        )}
      </Description>
    </dl>
  );
};

export default InfoPanel;
