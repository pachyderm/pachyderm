import classNames from 'classnames';
import React from 'react';

import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {
  Group,
  Icon,
  PachCronSVG,
  NavigationHistorySVG,
} from '@pachyderm/components';

import styles from './GlobalIdDisclaimer.module.css';

type GlobalIdDisclaimerProps = {
  globalId?: string;
  startTime?: string;
  artifact?: 'commit' | 'subjob';
};

const GlobalIdDisclaimer: React.FC<GlobalIdDisclaimerProps> = ({
  globalId,
  startTime,
  artifact,
}) => {
  return (
    <Group
      className={classNames(styles.borderBox, {
        [styles.active]: globalId,
      })}
      spacing={4}
    >
      <Icon className={styles.clockIcon}>
        {globalId ? <NavigationHistorySVG /> : <PachCronSVG />}
      </Icon>
      {globalId
        ? `Viewing: ${getStandardDateFromISOString(startTime)}`
        : `You are viewing the most recent ${artifact}`}
    </Group>
  );
};

export default GlobalIdDisclaimer;
