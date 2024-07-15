import React from 'react';

import {
  Icon,
  Button,
  FilterSVG,
  StatusCheckmarkSVG,
  StatusWarningSVG,
  ChipRadio,
} from '@pachyderm/components';

import {
  FILTER,
  FILTER_DAY,
  FILTER_HOUR,
  FILTER_WEEK,
} from '../../hooks/useGlobalFilter';

import styles from './SimpleDateFilters.module.css';

export const SimpleDateFilters = ({
  chips,
  showClearButton,
  clearFilters,
  toggleDateTimePicker,
  canHaveMoreJobs,
}: {
  chips: {
    lastHour: number;
    lastDay: number;
    lastWeek: number;
  };
  showClearButton: boolean;
  clearFilters: () => void;
  toggleDateTimePicker: () => void;
  canHaveMoreJobs: boolean;
}) => {
  const maxLastValue = Math.max(chips.lastHour, chips.lastDay, chips.lastWeek);
  return (
    <div className={styles.chipWrapper}>
      {!!chips.lastHour || !!chips.lastDay || !!chips.lastWeek ? (
        <div className={styles.chipContainer}>
          {!!chips.lastHour && (
            <ChipRadio
              name={FILTER}
              id={FILTER_HOUR}
              value={FILTER_HOUR}
              IconSVG={StatusWarningSVG}
            >
              {`(${
                canHaveMoreJobs && maxLastValue === chips.lastHour ? '≥' : ''
              }${chips.lastHour}) last hour`}
            </ChipRadio>
          )}
          {!!chips.lastDay && (
            <ChipRadio
              name={FILTER}
              value={FILTER_DAY}
              id={FILTER_DAY}
              IconSVG={StatusWarningSVG}
            >
              {`(${
                canHaveMoreJobs && maxLastValue === chips.lastDay ? '≥' : ''
              }${chips.lastDay}) last day`}
            </ChipRadio>
          )}
          {!!chips.lastWeek && (
            <ChipRadio
              name={FILTER}
              value={FILTER_WEEK}
              id={FILTER_WEEK}
              IconSVG={StatusWarningSVG}
            >
              {`(${
                canHaveMoreJobs && maxLastValue === chips.lastWeek ? '≥' : ''
              }${chips.lastWeek}) last 7 days`}
            </ChipRadio>
          )}
        </div>
      ) : (
        <span className={styles.noFailedJobsText}>
          <Icon color="green" small>
            <StatusCheckmarkSVG />
          </Icon>
          No failed jobs in the last week
        </span>
      )}
      <div className={styles.filterButtonGroup}>
        {showClearButton && (
          <Button onClick={clearFilters} buttonType="ghost">
            Clear
          </Button>
        )}
        <Button onClick={toggleDateTimePicker} buttonType="secondary">
          Filter
          <Icon color="plum" small>
            <FilterSVG />
          </Icon>
        </Button>
      </div>
    </div>
  );
};
