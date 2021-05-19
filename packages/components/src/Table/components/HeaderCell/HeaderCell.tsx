import classnames from 'classnames';
import noop from 'lodash/noop';
import React, {ThHTMLAttributes, useMemo} from 'react';

import {Group} from './../../../Group';
import {ArrowSVG} from './../../../Svg';
import styles from './HeaderCell.module.css';

export interface HeaderCellProps
  extends ThHTMLAttributes<HTMLTableHeaderCellElement> {
  rightAligned?: boolean;
  onClick?: () => void;
  sortable?: boolean;
  sortLabel?: string;
  sortSelected?: boolean;
  sortReversed?: boolean;
}

const HeaderCell: React.FC<HeaderCellProps> = ({
  children,
  onClick = noop,
  rightAligned = false,
  sortable = false,
  sortLabel = 'field',
  sortReversed = false,
  sortSelected = false,
  ...rest
}) => {
  const className = classnames(styles.base, {
    [styles.rightAligned]: rightAligned,
    [styles.sortHeader]: sortable,
  });

  const sortClasses = classnames({
    [styles.selectedColumn]: sortSelected,
    [styles.flip]: sortSelected && sortReversed,
  });

  const sortDir = useMemo(() => {
    if (!sortSelected) return 'none';
    if (sortReversed) return 'ascending';
    return 'descending';
  }, [sortReversed, sortSelected]);

  const ariaLabel = useMemo(
    () =>
      `sort by ${sortLabel} in ${
        sortDir !== 'descending' ? 'descending' : 'ascending'
      } order`,
    [sortDir, sortLabel],
  );

  return (
    <th
      className={className}
      {...rest}
      role="columnheader"
      aria-sort={sortDir}
      tab-index={0}
    >
      {sortable ? (
        <button
          aria-label={ariaLabel}
          className={styles.wrapperButton}
          onClick={onClick}
        >
          <Group spacing={8} align="center">
            {children}
            <ArrowSVG className={sortClasses} />
          </Group>
        </button>
      ) : (
        children
      )}
    </th>
  );
};

export default HeaderCell;
