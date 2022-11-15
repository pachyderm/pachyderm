import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import styles from './Row.module.css';
export interface RowProps extends HTMLAttributes<HTMLTableRowElement> {
  active?: boolean;
  sticky?: boolean;
}

const Row: React.FC<RowProps> = ({
  children,
  className,
  active = false,
  sticky = false,
  onClick,
  ...rest
}) => (
  <tr
    className={classnames(className, {
      [styles.active]: active,
      [styles.sticky]: sticky,
      [styles.button]: Boolean(onClick),
    })}
    role={onClick ? 'button' : undefined}
    onClick={onClick}
    {...rest}
  >
    {children}
  </tr>
);

export default Row;
