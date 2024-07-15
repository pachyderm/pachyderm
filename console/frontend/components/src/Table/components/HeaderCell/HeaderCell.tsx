import classnames from 'classnames';
import noop from 'lodash/noop';
import React, {ThHTMLAttributes, useMemo} from 'react';

import {
  Icon,
  ArrowDownSVG,
  ArrowUpSVG,
  CaptionText,
} from '@pachyderm/components';

import styles from './HeaderCell.module.css';

export interface HeaderCellProps
  extends ThHTMLAttributes<HTMLTableHeaderCellElement> {
  rightAligned?: boolean;
  onClick?: () => void;
  sortable?: boolean;
  sortLabel?: string;
  sortSelected?: boolean;
  sortReversed?: boolean;
  boxShadowBottomOnly?: boolean;
}

const HeaderCell: React.FC<HeaderCellProps> = ({
  children,
  onClick = noop,
  rightAligned = false,
  sortable = false,
  sortLabel = 'field',
  sortReversed = false,
  sortSelected = false,
  boxShadowBottomOnly = false,
  className,
  ...rest
}) => {
  const classNames = classnames(className, styles.base, {
    [styles.rightAligned]: rightAligned,
    [styles.sortHeader]: sortable,
    [styles.boxShadowBottomOnly]: boxShadowBottomOnly,
    className,
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
      className={classNames}
      {...rest}
      role="columnheader"
      aria-sort={sortDir}
      tabIndex={0}
    >
      {sortable ? (
        <button
          aria-label={ariaLabel}
          className={styles.wrapperButton}
          onClick={onClick}
          type="button"
        >
          <CaptionText className={styles.text}>{children}</CaptionText>
          <Icon color={sortSelected ? 'blue' : 'grey'} small>
            {sortSelected && sortReversed ? <ArrowDownSVG /> : <ArrowUpSVG />}
          </Icon>
        </button>
      ) : (
        <CaptionText className={styles.text}>{children}</CaptionText>
      )}
    </th>
  );
};

export default HeaderCell;
