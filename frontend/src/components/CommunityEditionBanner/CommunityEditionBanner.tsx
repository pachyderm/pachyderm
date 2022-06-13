import {Button} from '@pachyderm/components';
import classnames from 'classnames';
import {formatDistanceToNowStrict, fromUnixTime} from 'date-fns';
import React from 'react';

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
          ? `Access ends in ${formatDistanceToNowStrict(
              fromUnixTime(expiration),
            )}`
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
          href="http://pachyderm.com/trial-console"
          buttonType="secondary"
        >
          Upgrade to Enterprise
        </Button>
      </div>
    </div>
  );
};

export default CommunityEditionBanner;
