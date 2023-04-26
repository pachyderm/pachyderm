import classnames from 'classnames';
import React from 'react';

import {BrandedErrorIcon} from '@dash-frontend/components/BrandedIcon';
import {Group} from '@pachyderm/components';

import styles from './GenericError.module.css';

export const GenericError: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
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

    <BrandedErrorIcon className={styles.svg} disableDefaultStyling />
  </Group>
);

export default GenericError;
