import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import styles from './Body.module.css';

export interface BodyProps extends HTMLAttributes<HTMLTableSectionElement> {
  bandedRows?: boolean;
}

const Body: React.FC<BodyProps> = ({
  children,
  className,
  bandedRows = true,
  ...rest
}) => (
  <tbody
    className={classnames(className, {[styles.bandedRows]: bandedRows})}
    {...rest}
  >
    {children}
  </tbody>
);

export default Body;
