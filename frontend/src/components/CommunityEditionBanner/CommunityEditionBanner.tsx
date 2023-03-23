import classnames from 'classnames';
import React from 'react';

import {getDurationToNow} from '@dash-frontend/lib/dateTime';
import {Button} from '@pachyderm/components';

import styles from './CommunityEditionBanner.module.css';
import useCommunityEditionBanner, {
  PIPELINE_LIMIT,
} from './useCommunityEditionBanner';

type CommunityEditionBannerProps = {
  expiration?: number;
};

const CommunityEditionBanner: React.FC<CommunityEditionBannerProps> = ({
  expiration,
}) => {
  const {pipelines, pipelineLimitReached, workerLimitReached} =
    useCommunityEditionBanner(expiration);

  return (
    <div
      className={classnames(styles.communityEditionBanner, {
        [styles.limitReached]: pipelineLimitReached || workerLimitReached,
      })}
    >
      <div className={styles.license}>
        <strong>{expiration ? 'Enterprise Key' : 'Community Edition'}</strong>
      </div>
      <div className={styles.limits}>
        {expiration
          ? `Access ends in ${getDurationToNow(expiration)}`
          : `Pipelines: ${pipelines?.length || 0}/${PIPELINE_LIMIT}`}
      </div>

      <div className={styles.rightAlign}>
        <div className={styles.limitMessage}>
          {pipelineLimitReached ? 'Reaching pipeline limit' : null}
          {!pipelineLimitReached && workerLimitReached
            ? 'Reaching worker limit (8 per pipeline)'
            : null}
        </div>
        <Button
          href="https://www.pachyderm.com/trial-console/?utm_source=console"
          buttonType="secondary"
          data-testid="CommunityEditionBanner__enterpriseUpgrade"
        >
          Upgrade to Enterprise
        </Button>
      </div>
    </div>
  );
};

export default CommunityEditionBanner;
