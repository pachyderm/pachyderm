import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {Dispatch, SetStateAction} from 'react';

import {Search, Group} from '@pachyderm/components';

import styles from './BodyHeaderTabs.module.css';

export type BodyHeaderTabsProps = {
  children?: React.ReactNode;
  searchValue?: string;
  onSearch?: Dispatch<SetStateAction<string>>;
  placeholder?: string;
  showSearch?: boolean;
  className?: string;
};

const BodyHeaderTabs: React.FC<BodyHeaderTabsProps> = ({
  searchValue = '',
  onSearch = noop,
  placeholder = 'Search',
  showSearch = false,
  className,
  children,
}) => {
  return (
    <Group spacing={32} className={styles.base}>
      {children}
      {showSearch && (
        <Search
          value={searchValue}
          className={classNames(styles.search, className)}
          placeholder={placeholder}
          onSearch={onSearch}
          aria-label={placeholder}
        />
      )}
    </Group>
  );
};

export default BodyHeaderTabs;
