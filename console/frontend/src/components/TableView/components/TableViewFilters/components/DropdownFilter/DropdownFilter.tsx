/* eslint-disable @typescript-eslint/no-explicit-any */
import React, {useState} from 'react';
import {UseFormReturn} from 'react-hook-form';

import {
  Checkbox,
  Dropdown,
  Button,
  SearchSVG,
  CloseSVG,
  Icon,
} from '@pachyderm/components';

import styles from './DropdownFilter.module.css';

type DropdownFilterProps = {
  formCtx: UseFormReturn<any>;
  multiselectFilter: MultiselectFilter;
};

export type MultiselectFilter = {
  label: string;
  noun: string;
  name: string;
  formatLabel?: (val: string) => string;
  values: string[];
};

const shorten = (val: string, limit: number) => {
  if (val.length <= limit) {
    return val;
  }
  return `${val.slice(0, limit)}...`;
};

const DropdownFilter: React.FC<DropdownFilterProps> = ({
  formCtx,
  multiselectFilter,
}) => {
  const [searchFilter, setSearchFilter] = useState('');
  const {getValues} = formCtx;
  const filterValues = getValues(multiselectFilter.name);
  return (
    <Dropdown formCtx={formCtx} className={styles.base}>
      <Dropdown.Button className={styles.dropdownFilterButton}>
        {filterValues.length === 0 && `Select ${multiselectFilter.noun}s`}
        {filterValues.length > 0 && `${shorten(filterValues[0], 6)}`}
        {filterValues.length > 1 && `, ${shorten(filterValues[1], 6)}`}
        {filterValues.length > 2 && `, +${filterValues.length - 2}`}
      </Dropdown.Button>
      <Dropdown.Menu className={styles.menu}>
        <div className={styles.search}>
          <Icon className={styles.searchIcon}>
            <SearchSVG aria-hidden />
          </Icon>

          <input
            className={styles.input}
            placeholder={`Enter exact ${multiselectFilter.noun}`}
            onChange={(e) => setSearchFilter(e.target.value)}
            value={searchFilter}
          />

          {searchFilter && (
            <Button
              aria-label="Clear"
              buttonType="ghost"
              onClick={() => setSearchFilter('')}
              IconSVG={CloseSVG}
            />
          )}
        </div>
        {multiselectFilter.values
          .filter((value) => !searchFilter || value === searchFilter)
          .map((value) => (
            <Dropdown.MenuItem key={value} id={value}>
              <Checkbox
                small
                id={value}
                value={value}
                name={multiselectFilter.name}
                label={
                  multiselectFilter.formatLabel
                    ? multiselectFilter.formatLabel(value)
                    : value
                }
              />
            </Dropdown.MenuItem>
          ))}
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default DropdownFilter;
