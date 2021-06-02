import {fromUnixTime, formatDistanceToNow, formatDistance} from 'date-fns';
import React, {useMemo} from 'react';
import {useRouteMatch} from 'react-router';

import Description from '@dash-frontend/components/Description';
import JSONBlock from '@dash-frontend/components/JSONBlock';
import PipelineInput from '@dash-frontend/components/PipelineInput';
import {ProjectRouteParams} from '@dash-frontend/lib/types';

import pipelineJobs from '../../mock/pipelineJobs';

import styles from './InfoPanel.module.css';

interface InfoPanelParams extends ProjectRouteParams {
  pipelineJobId: string;
}

const InfoPanel = () => {
  const {
    params: {pipelineJobId, projectId},
  } = useRouteMatch<InfoPanelParams>();

  // TODO: Replace with `usePipelineJob`
  const pipelineJob = pipelineJobs.find((job) => job.id === pipelineJobId);

  const transformString = useMemo(() => {
    if (pipelineJob?.transform) {
      return (
        <JSONBlock>
          {JSON.stringify(
            {
              image: pipelineJob.transform.image,
              cmd: pipelineJob.transform.cmdList,
            },
            null,
            2,
          )}
        </JSONBlock>
      );
    }
    return 'N/A';
  }, [pipelineJob?.transform]);

  const started = useMemo(() => {
    return pipelineJob?.createdAt
      ? formatDistanceToNow(fromUnixTime(pipelineJob.createdAt), {
          addSuffix: true,
        })
      : 'N/A';
  }, [pipelineJob?.createdAt]);

  const duration = useMemo(() => {
    return pipelineJob?.finishedAt && pipelineJob?.createdAt
      ? formatDistance(
          fromUnixTime(pipelineJob.finishedAt),
          fromUnixTime(pipelineJob.createdAt),
        )
      : 'N/A';
  }, [pipelineJob?.createdAt, pipelineJob?.finishedAt]);

  return (
    <dl className={styles.base}>
      <Description term="Inputs" lines={9}>
        {pipelineJob?.inputString && (
          <PipelineInput
            inputString={pipelineJob.inputString}
            branchId={pipelineJob.inputBranch || 'master'}
            projectId={projectId}
          />
        )}
      </Description>

      <Description term="Transform">{transformString}</Description>

      <hr className={styles.divider} />

      <Description term="ID">{pipelineJob?.id}</Description>
      <Description term="Pipeline">{pipelineJob?.pipelineName}</Description>
      <Description term="Started">{started}</Description>
      <Description term="Duration">{duration}</Description>
    </dl>
  );
};

export default InfoPanel;
