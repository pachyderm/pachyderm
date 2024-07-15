import classNames from 'classnames';
import React from 'react';

import {Group, CaptionTextSmall} from '@pachyderm/components';

import styles from './Info.module.css';

type InfoProps = {
  children?: React.ReactNode;
  header: string;
  headerId: string;
  className?: string;
  disabled?: boolean;
  wide?: boolean;
};

export const Info: React.FC<InfoProps> = ({
  header,
  headerId,
  children,
  className,
  disabled,
  wide = false,
}) => {
  return (
    <Group vertical spacing={8} className={className}>
      <CaptionTextSmall id={headerId}>{header}</CaptionTextSmall>
      <span
        aria-labelledby={headerId}
        className={classNames(wide ? styles.info40 : styles.info20, {
          [styles.disabled]: disabled,
        })}
      >
        {children}
      </span>
    </Group>
  );
};

export default Info;
