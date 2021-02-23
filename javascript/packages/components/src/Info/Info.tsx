import classNames from 'classnames';
import React from 'react';

import {Group} from 'Group';

import styles from './Info.module.css';

type InfoProps = {
  header: string;
  headerId: string;
  className?: string;
  disabled?: boolean;
};

const Info: React.FC<InfoProps> = ({
  header,
  headerId,
  children,
  className,
  disabled,
}) => {
  return (
    <Group vertical spacing={8} className={className}>
      <span className={styles.header} id={headerId}>
        {header}
      </span>
      <span
        aria-labelledby={headerId}
        className={classNames(styles.info, {[styles.disabled]: disabled})}
      >
        {children}
      </span>
    </Group>
  );
};

export default Info;
