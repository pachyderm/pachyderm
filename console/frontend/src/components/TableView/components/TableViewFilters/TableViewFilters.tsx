/* eslint-disable @typescript-eslint/no-explicit-any */
import capitalize from 'lodash/capitalize';
import React from 'react';
import {UseFormReturn} from 'react-hook-form';

import {ChipGroup, Chip} from '@dash-frontend/components/Chip';
import ExpandableSection from '@dash-frontend/components/ExpandableSection';
import {
  Button,
  CaptionTextSmall,
  RadioButton,
  Checkbox,
  CloseSVG,
} from '@pachyderm/components';

import DropdownFilter, {
  MultiselectFilter,
} from './components/DropdownFilter/DropdownFilter';
import styles from './TableViewFilters.module.css';

export const MAX_FILTER_HEIGHT_REM = 12;

type optionType = {
  name: string;
  value: string;
};

type filterType = {
  label: string;
  name: string;
  type: string;
  options: optionType[];
};

type TableViewFiltersProps = {
  formCtx: UseFormReturn<any>;
  filtersExpanded: boolean;
  clearableFiltersMap?: {
    field: string;
    name: string;
    value: string;
  }[];
  staticFilterKeys?: string[];
  filters?: filterType[];
  multiselectFilters?: MultiselectFilter[];
};

const getInputType = (filter: filterType, option: optionType) => {
  switch (filter.type) {
    case 'checkbox':
      return (
        <Checkbox
          key={option.value}
          id={option.value}
          value={option.value}
          name={filter.name}
          label={option.name}
          small
        />
      );
    case 'radio':
      return (
        <RadioButton
          key={option.value}
          id={option.value}
          name={filter.name}
          value={option.value}
          small
        >
          <RadioButton.Label>{option.name}</RadioButton.Label>
        </RadioButton>
      );
    default:
      return null;
  }
};

const TableViewFilters: React.FC<TableViewFiltersProps> = ({
  formCtx,
  filtersExpanded,
  clearableFiltersMap = [],
  staticFilterKeys = [],
  filters = [],
  multiselectFilters = [],
}) => {
  const {setValue} = formCtx;
  return (
    <>
      <ExpandableSection
        expanded={filtersExpanded}
        maxHeightRem={MAX_FILTER_HEIGHT_REM}
      >
        {filters.map((section) => (
          <div className={styles.filterSection} key={section.label}>
            <CaptionTextSmall className={styles.filterLabel}>
              {section.label}
            </CaptionTextSmall>
            {section.options.map((option) => getInputType(section, option))}
          </div>
        ))}
        {multiselectFilters.map((multiselectFilter) => (
          <div className={styles.filterSection} key={multiselectFilter.label}>
            <CaptionTextSmall className={styles.filterLabel}>
              {multiselectFilter.label}
            </CaptionTextSmall>
            <DropdownFilter
              formCtx={formCtx}
              multiselectFilter={multiselectFilter}
            />
          </div>
        ))}
      </ExpandableSection>
      {!(clearableFiltersMap.length === 0 && staticFilterKeys.length === 0) && (
        <ChipGroup className={styles.chipGroup}>
          {clearableFiltersMap.map(({field, name, value}) => {
            // prepare an array of filter values without the given filter
            // if the user clicks on the chip to remove it
            const filterOmitted = clearableFiltersMap
              .filter(
                (filter) => filter.field === field && filter.value !== value,
              )
              .map((filter) => filter.value);
            return (
              <Chip
                data-testid={`Filter__${field}${name}Chip`}
                key={`${field}${name}`}
                onClick={() => setValue(field, filterOmitted)}
                RightIconSVG={CloseSVG}
              >
                {capitalize(name)}
              </Chip>
            );
          })}
          {staticFilterKeys.map((item) => (
            <Chip
              data-testid={`Filter__${item}Chip`}
              key={item}
              isButton={false}
            >
              {item}
            </Chip>
          ))}
          {clearableFiltersMap.length > 0 && (
            <Button buttonType="ghost" onClick={() => formCtx.reset()}>
              Clear
            </Button>
          )}
        </ChipGroup>
      )}
    </>
  );
};

export default TableViewFilters;
