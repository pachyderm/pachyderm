import React, {useCallback} from 'react';
import {useHistory} from 'react-router';

import {DatumState, JobInfo} from '@dash-frontend/api/pps';
import IconBadge from '@dash-frontend/components/IconBadge';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DatumFilter} from '@dash-frontend/lib/types';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  StatusCheckmarkSVG,
  StatusWarningSVG,
  StatusSkipSVG,
  StatusUpdatedSVG,
  DropdownItem,
  CaptionTextSmall,
} from '@pachyderm/components';

import styles from '../../JobSetList/components/RunsTable/components/RunsList/RunsList.module.css';

type DatumBadgeProps = {
  count?: string;
  tooltip: string;
  color: 'red' | 'green' | 'black';
  IconSVG: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  filter: DatumFilter;
};

const useRunsList = () => {
  const {projectId} = useUrlState();
  const {getPathToJobLogs, getPathToDatumLogs} = useLogsNavigation();
  const {getUpdatedSearchParams} = useUrlQueryState();
  const browserHistory = useHistory();

  const globalIdRedirect = (runId: string) => {
    const searchParams = getUpdatedSearchParams(
      {
        globalIdFilter: runId,
      },
      true,
    );

    return browserHistory.push(
      `${lineageRoute(
        {
          projectId,
        },
        false,
      )}?${searchParams}`,
    );
  };

  const inspectJobsRedirect = (jobId: string, pipelineId: string) => {
    const logsLink = getPathToJobLogs({
      projectId,
      jobId: jobId,
      pipelineId: pipelineId,
    });
    browserHistory.push(logsLink);
  };

  const onOverflowMenuSelect =
    (runId: string, pipelineId: string) => (id: string) => {
      switch (id) {
        case 'apply-run':
          return globalIdRedirect(runId);
        case 'inspect-job':
          return inspectJobsRedirect(runId, pipelineId);
        default:
          return null;
      }
    };

  const iconItems: DropdownItem[] = [
    {
      id: 'inspect-job',
      content: 'Inspect job',
      closeOnClick: true,
    },
    {
      id: 'apply-run',
      content: 'Apply Global ID and view in DAG',
      closeOnClick: true,
    },
  ];

  const getDatumStateBadges = useCallback(
    (job: JobInfo) => {
      const Badge: React.FC<DatumBadgeProps> = ({
        count,
        tooltip,
        color,
        IconSVG,
        filter,
      }) => {
        const datumLogsLink = (
          jobId: string,
          pipelineId: string,
          datumFilter: DatumFilter,
        ) => {
          return getPathToDatumLogs(
            {
              projectId,
              jobId: jobId,
              pipelineId: pipelineId,
            },
            [datumFilter],
          );
        };

        return Number(count) > 0 ? (
          <IconBadge
            aria-label={tooltip}
            color={color}
            IconSVG={IconSVG}
            tooltip={
              <>
                <p>{tooltip}</p>
                <CaptionTextSmall>Click to inspect</CaptionTextSmall>
              </>
            }
            to={datumLogsLink(
              job?.job?.id || '',
              job?.job?.pipeline?.name || '',
              filter,
            )}
          >
            {count}
          </IconBadge>
        ) : null;
      };

      return (
        <span className={styles.jobStates}>
          <Badge
            count={job.dataProcessed}
            color="green"
            IconSVG={StatusCheckmarkSVG}
            tooltip="Processed datums"
            filter={DatumState.SUCCESS}
          />
          <Badge
            count={job.dataFailed}
            color="red"
            IconSVG={StatusWarningSVG}
            tooltip="Failed datums"
            filter={DatumState.FAILED}
          />
          <Badge
            count={job.dataSkipped}
            color="black"
            IconSVG={StatusSkipSVG}
            tooltip="Skipped datums"
            filter={DatumState.SKIPPED}
          />
          <Badge
            count={job.dataRecovered}
            color="black"
            IconSVG={StatusUpdatedSVG}
            tooltip="Recovered datums"
            filter={DatumState.RECOVERED}
          />
        </span>
      );
    },
    [getPathToDatumLogs, projectId],
  );

  return {
    iconItems,
    onOverflowMenuSelect,
    getDatumStateBadges,
  };
};

export default useRunsList;
