import classNames from 'classnames';
import React from 'react';

import {Group, CaptionTextSmall} from '@pachyderm/components';

import styles from './Info.module.css';

type InfoProps = {
  header: string;
  headerId: string;
  className?: string;
  disabled?: boolean;
};

export const Info: React.FC<InfoProps> = ({
  header,
  headerId,
  children,
  className,
  disabled,
}) => {
  return (
    <Group vertical spacing={8} className={className}>
      <CaptionTextSmall id={headerId}>{header}</CaptionTextSmall>
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
