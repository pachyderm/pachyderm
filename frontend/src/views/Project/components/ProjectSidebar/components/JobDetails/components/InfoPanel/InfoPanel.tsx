import {Link} from '@pachyderm/components';
import {
  fromUnixTime,
  formatDistanceToNow,
  formatDistanceStrict,
} from 'date-fns';
import React, {useMemo} from 'react';

import Description from '@dash-frontend/components/Description';
import JSONBlock from '@dash-frontend/components/JSONBlock';
import JsonSpec from '@dash-frontend/components/JsonSpec';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import extractAndShortenIds from '@dash-frontend/lib/extractAndShortenIds';
import readableJobState from '@dash-frontend/lib/readableJobState';
import ReadLogsButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/ReadLogsButton';

import styles from './InfoPanel.module.css';

type InfoPanelProps = {
  showReadLogs?: boolean;
};

const InfoPanel: React.FC<InfoPanelProps> = ({showReadLogs = false}) => {
  const {jobId, projectId, pipelineId} = useUrlState();
  const {viewState} = useUrlQueryState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();

  const {job, loading} = useJob({
    id: viewState.globalIdFilter || jobId,
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

  return (
    <dl className={styles.base}>
      {showReadLogs && (
        <div className={styles.logsButtonWrapper}>
          <ReadLogsButton />
        </div>
      )}
      <Description
        term="Output Commit"
        loading={loading}
        data-testid="InfoPanel__id"
      >
        {job?.outputBranch ? (
          <Link
            to={getPathToFileBrowser({
              projectId,
              commitId: job.id,
              branchId: job.outputBranch,
              repoId: job.pipelineName,
            })}
            data-testid="InfoPanel__commitLink"
          >
            {job?.id}
          </Link>
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
