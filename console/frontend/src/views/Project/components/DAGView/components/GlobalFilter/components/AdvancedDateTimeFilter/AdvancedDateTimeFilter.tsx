import React from 'react';
import {useFormContext} from 'react-hook-form';

import {Icon, Button, FilterSVG, ArrowRightSVG} from '@pachyderm/components';

import styles from './AdvancedDateTimeFilter.module.css';

export const AdvancedDateTimeFilter = ({
  dateTimeFilterValue,
  toggleDateTimePicker,
}: {
  dateTimeFilterValue: string | null;
  toggleDateTimePicker: () => void;
}) => {
  const {register} = useFormContext();

  return (
    <>
      <div className={styles.chipWrapper}>
        <div className={styles.chipContainer}>
          {dateTimeFilterValue && (
            <div className={styles.dateTimeValue}>{dateTimeFilterValue}</div>
          )}
        </div>
        <Button onClick={toggleDateTimePicker} buttonType="secondary">
          Filter
          <Icon color="plum" small>
            <FilterSVG />
          </Icon>
        </Button>
      </div>
      <div className={styles.filtersContainer}>
        <div className={styles.dateTimeRow}>
          <input
            className={styles.dateTimeInput}
            type="date"
            id="startDate"
            data-testid="AdvancedDateTimeFilter__startDate"
            aria-label="Start date"
            {...register('startDate')}
          />

          <Icon small color="grey">
            <ArrowRightSVG />
          </Icon>

          <input
            className={styles.dateTimeInput}
            type="date"
            id="endDate"
            data-testid="AdvancedDateTimeFilter__endDate"
            aria-label="End date"
            {...register('endDate')}
          />
        </div>
        <div className={styles.dateTimeRow}>
          <input
            className={styles.dateTimeInput}
            type="time"
            id="startTime"
            data-testid="AdvancedDateTimeFilter__startTime"
            aria-label="Start time"
            {...register('startTime')}
          />

          <Icon small color="grey">
            <ArrowRightSVG />
          </Icon>

          <input
            className={styles.dateTimeInput}
            type="time"
            id="endTime"
            data-testid="AdvancedDateTimeFilter__endTime"
            aria-label="End time"
            {...register('endTime')}
          />
        </div>
      </div>
    </>
  );
};
