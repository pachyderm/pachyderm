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
          {/* <label htmlFor="startDate">Start date</label> */}
          <input
            className={styles.dateTimeInput}
            type="date"
            id="startDate"
            data-testid="AdvancedDateTimeFilter__startDate"
            {...register('startDate')}
          />

          <Icon small color="grey">
            <ArrowRightSVG />
          </Icon>

          {/* <label htmlFor="endDate">End date</label> */}
          <input
            className={styles.dateTimeInput}
            type="date"
            id="endDate"
            data-testid="AdvancedDateTimeFilter__endDate"
            {...register('endDate')}
          />
        </div>
        <div className={styles.dateTimeRow}>
          {/* <label htmlFor="startTime">Start Time</label> */}
          <input
            className={styles.dateTimeInput}
            type="time"
            id="startTime"
            data-testid="AdvancedDateTimeFilter__startTime"
            {...register('startTime')}
          />

          <Icon small color="grey">
            <ArrowRightSVG />
          </Icon>

          {/* <label htmlFor="endTime">End Time</label> */}
          <input
            className={styles.dateTimeInput}
            type="time"
            id="endTime"
            data-testid="AdvancedDateTimeFilter__endTime"
            {...register('endTime')}
          />
        </div>
      </div>
    </>
  );
};
