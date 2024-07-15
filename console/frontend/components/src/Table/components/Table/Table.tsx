import classnames from 'classnames';
import React, {TableHTMLAttributes} from 'react';

import styles from './Table.module.css';

export interface TableProps extends TableHTMLAttributes<HTMLTableElement> {
  caption?: string;
}

const Table: React.FC<TableProps> = ({
  caption,
  children,
  className,
  ...rest
}) => (
  <table className={classnames(styles.base, className)} {...rest}>
    {caption && <caption className="visually-hide">{caption}</caption>}
    {children}
  </table>
);

export default Table;
