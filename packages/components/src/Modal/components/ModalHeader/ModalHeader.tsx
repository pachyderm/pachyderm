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
  withStatus?: boolean;
}

const ModalHeader: React.FC<ModalHeaderProps> = ({
  children,
  onHide,
  actionable = false,
  small = false,
  withStatus = false,
  ref, // Note: The ModalHeader from Bootstrap errors out when forwarding a ref
  ...props
}) => {
  return (
    <BootstrapModalHeader
      {...props}
      className={classNames(styles.base, {
        [styles.small]: small,
        [styles.withStatus]: withStatus,
      })}
    >
      <Group spacing={8} align="center">
        <h4>{children}</h4>
      </Group>
    </BootstrapModalHeader>
  );
};

export default ModalHeader;
