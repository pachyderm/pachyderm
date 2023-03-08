import {DatumFilter} from '@graphqlTypes';
import classnames from 'classnames';
import capitalize from 'lodash/capitalize';
import xor from 'lodash/xor';
import React from 'react';
import {UseFormReturn} from 'react-hook-form';

import {Chip, ChipGroup} from '@dash-frontend/components/Chip';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {readableDatumState} from '@dash-frontend/lib/datums';
import {
  CaptionTextSmall,
  CloseSVG,
  Dropdown,
  FilterSVG,
  Form,
  PureCheckbox,
  RadioButton,
} from '@pachyderm/components';

import {DatumFilterFormValues} from '../../hooks/useLeftPanel';

import styles from './Filter.module.css';

export const datumFilters = Object.values(DatumFilter);

export type FilterProps = {
  formCtx: UseFormReturn<DatumFilterFormValues>;
};

export type stateOptions = {
  id: DatumFilter;
  name: string;
};

export const Filter: React.FC<FilterProps> = ({formCtx}) => {
  const {searchParams, updateSearchParamsAndGo} = useUrlQueryState();
  const filters = searchParams.datumFilters || [];

  const {watch, setValue} = formCtx;
  const jobsSort = watch('jobs');

  const resetJobFilter = () => {
    if (jobsSort === 'status') {
      setValue('jobs', 'newest');
    }
  };

  const updateDatumSelection = (datumState: DatumFilter) => {
    updateSearchParamsAndGo({
      datumFilters: xor<DatumFilter>(filters as DatumFilter[], [datumState]),
    });
  };

  return (
    <div className={styles.base}>
      <Dropdown>
        <Dropdown.Button
          className={styles.dropdownButton}
          IconSVG={FilterSVG}
          iconPosition="start"
          color="plumb"
          buttonType="ghost"
        >
          Filter
        </Dropdown.Button>
        <Dropdown.Menu pin="left" className={styles.dropdownMenu}>
          <Form formContext={formCtx}>
            <div className={classnames(styles.divider, styles.group)}>
              <div className={styles.heading}>
                <CaptionTextSmall>Sort Jobs By</CaptionTextSmall>
              </div>
              <RadioButton id="newest" name="jobs" value="newest">
                <RadioButton.Label>Newest</RadioButton.Label>
              </RadioButton>
              <RadioButton id="status" name="jobs" value="status">
                <RadioButton.Label>Job status</RadioButton.Label>
              </RadioButton>
            </div>
            <div className={styles.group}>
              <div className={styles.heading}>
                <CaptionTextSmall>Filter Datums by Status</CaptionTextSmall>
              </div>
              {datumFilters.map((state) => (
                <PureCheckbox
                  small
                  id={state}
                  key={state}
                  name={state}
                  label={readableDatumState(state)}
                  onClick={() => updateDatumSelection(state)}
                  selected={filters.includes(state)}
                />
              ))}
            </div>
          </Form>
        </Dropdown.Menu>
      </Dropdown>
      <ChipGroup>
        <Chip
          onClick={resetJobFilter}
          RightIconSVG={jobsSort === 'status' ? CloseSVG : undefined}
        >
          {capitalize(jobsSort)}
        </Chip>
        {filters &&
          filters.map((item) => (
            <Chip
              data-testid={`Filter__${item}Chip`}
              key={item}
              onClick={() => updateDatumSelection(item as DatumFilter)}
              RightIconSVG={CloseSVG}
            >
              {readableDatumState(item)}
            </Chip>
          ))}
      </ChipGroup>
    </div>
  );
};

export default Filter;
