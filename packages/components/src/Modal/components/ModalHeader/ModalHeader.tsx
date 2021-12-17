import classNames from 'classnames';
import React from 'react';
import BootstrapModalHeader, {
  ModalHeaderProps as BootstrapModalHeaderProps,
} from 'react-bootstrap/ModalHeader';

import {Group} from './../../../Group';
import styles from './ModalHeader.module.css';

export interface ModalHeaderProps
  extends Omit<BootstrapModalHeaderProps, 'onHide'> {
  actionable?: boolean;
  onHide: () => void;
  small?: boolean;
}

const ModalHeader: React.FC<ModalHeaderProps> = ({
  children,
  onHide,
  actionable = false,
  small = false,
  ref, // Note: The ModalHeader from Bootstrap errors out when forwarding a ref
  ...props
}) => {
  return (
    <BootstrapModalHeader
      {...props}
      className={classNames(styles.base, {
        [styles.small]: small,
      })}
    >
      <Group spacing={8} align="center">
        {children}
      </Group>
    </BootstrapModalHeader>
  );
};

export default ModalHeader;
