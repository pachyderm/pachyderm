import React from 'react';

import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import IconBadge from '@dash-frontend/components/IconBadge';
import {InternalJobSet, NodeState} from '@dash-frontend/lib/types';
import {
  StatusDotsSVG,
  StatusCheckmarkSVG,
  StatusWarningSVG,
} from '@pachyderm/components';

import styles from './JobStateBadges.module.css';

type JobStateBadgesProps = {
  jobSet: InternalJobSet;
  showToolTips?: boolean;
};

const JobStateBadges: React.FC<JobStateBadgesProps> = ({
  jobSet,
  showToolTips = true,
}) => {
  const nodeStates = jobSet.jobs.reduce(
    (acc, job) => {
      acc[restJobStateToNodeState(job.state)] += 1;
      return acc;
    },
    {
      [NodeState.ERROR]: 0,
      [NodeState.BUSY]: 0,
      [NodeState.IDLE]: 0,
      [NodeState.PAUSED]: 0,
      [NodeState.RUNNING]: 0,
      [NodeState.SUCCESS]: 0,
    },
  );

  return (
    <div className={styles.base}>
      {nodeStates[NodeState.RUNNING] > 0 && (
        <IconBadge
          color="green"
          IconSVG={StatusDotsSVG}
          tooltip={showToolTips && 'Running jobs'}
        >
          {nodeStates[NodeState.RUNNING]}
        </IconBadge>
      )}
      {nodeStates[NodeState.SUCCESS] > 0 && (
        <IconBadge
          color="green"
          IconSVG={StatusCheckmarkSVG}
          tooltip={showToolTips && 'Successful jobs'}
        >
          {nodeStates[NodeState.SUCCESS]}
        </IconBadge>
      )}
      {nodeStates[NodeState.ERROR] > 0 && (
        <IconBadge
          color="red"
          IconSVG={StatusWarningSVG}
          tooltip={showToolTips && 'Failed jobs'}
        >
          {nodeStates[NodeState.ERROR]}
        </IconBadge>
      )}
    </div>
  );
};

export default JobStateBadges;
