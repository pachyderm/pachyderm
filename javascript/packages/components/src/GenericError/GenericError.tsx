import classnames from 'classnames';
import React from 'react';

import {Group} from 'Group';
import {GenericErrorSVG} from 'Svg';

import styles from './GenericError.module.css';

const GenericError: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...rest
}) => (
  <Group
    vertical
    spacing={16}
    align="center"
    className={classnames(styles.base, className)}
    {...rest}
  >
    <span role="status" className={styles.text}>
      {children}
    </span>

    <GenericErrorSVG className={styles.svg} />
  </Group>
);

export default GenericError;
