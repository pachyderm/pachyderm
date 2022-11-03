import classnames from 'classnames';
import React, {TdHTMLAttributes} from 'react';

import styles from './DataCell.module.css';

export interface DataCellProps
  extends TdHTMLAttributes<HTMLTableDataCellElement> {
  rightAligned?: boolean;
}

const DataCell: React.FC<DataCellProps> = ({
  children,
  rightAligned = false,
  className,
  ...rest
}) => {
  const classes = classnames(styles.base, className, {
    [styles.rightAligned]: rightAligned,
  });

  return (
    <td className={classes} {...rest}>
      {children}
    </td>
  );
};

export default DataCell;
