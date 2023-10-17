import classnames from 'classnames';
import React, {TdHTMLAttributes} from 'react';

import useStickyState from '../../../hooks/useStickyState';

import styles from './DataCell.module.css';

export interface DataCellProps
  extends TdHTMLAttributes<HTMLTableDataCellElement> {
  rightAligned?: boolean;
  cellRef?: React.LegacyRef<HTMLTableDataCellElement>;
  sticky?: boolean;
  isSelected?: boolean;
}

const DataCell: React.FC<DataCellProps> = ({
  children,
  rightAligned = false,
  className,
  cellRef: _callRef,
  sticky = false,
  isSelected = false,
  ...rest
}) => {
  const {stickyRef, elementIsStuck} = useStickyState();

  const classes = classnames(
    styles.base,
    {
      [styles.rightAligned]: rightAligned,
      [styles.sticky]: sticky,
      [styles.stuck]: elementIsStuck,
      [styles.isSelected]: isSelected && elementIsStuck,
    },
    className,
  );

  return (
    <td className={classes} ref={sticky ? stickyRef : undefined} {...rest}>
      {children}
    </td>
  );
};

export default DataCell;
