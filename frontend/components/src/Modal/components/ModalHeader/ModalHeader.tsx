import classNames from 'classnames';
import React from 'react';

import {Group} from './../../../Group';
import styles from './ModalHeader.module.css';

type ModalHeaderProps = {
  small?: boolean;
  withStatus?: boolean;
  children?: React.ReactNode;
  className?: string;
};

const ModalHeader: React.FC<ModalHeaderProps> = ({
  children,
  small = false,
  withStatus = false,
  className,
}) => {
  return (
    <div
      className={classNames(
        styles.base,
        {
          [styles.small]: small,
          [styles.withStatus]: withStatus,
        },
        className,
      )}
    >
      <Group spacing={8} align="center">
        <h4>{children}</h4>
      </Group>
    </div>
  );
};

export default ModalHeader;
