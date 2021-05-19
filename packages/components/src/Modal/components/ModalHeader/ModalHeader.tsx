import classNames from 'classnames';
import React from 'react';
import BootstrapModalHeader, {
  ModalHeaderProps as BootstrapModalHeaderProps,
} from 'react-bootstrap/ModalHeader';
import Spinner from 'react-bootstrap/Spinner';

import {Group} from './../../../Group';
import {SuccessCheckmark} from './../../../SuccessCheckmark';
import styles from './ModalHeader.module.css';

export interface ModalHeaderProps
  extends Omit<BootstrapModalHeaderProps, 'onHide'> {
  actionable?: boolean;
  onHide: () => void;
  important?: boolean;
  success?: boolean;
  error?: boolean;
  loading?: boolean;
}

const ModalHeader: React.FC<ModalHeaderProps> = ({
  children,
  onHide,
  actionable = false,
  important = false,
  success = false,
  error = false,
  loading = false,
  ref, // Note: The ModalHeader from Bootstrap errors out when forwarding a ref
  ...props
}) => {
  return (
    <BootstrapModalHeader
      {...props}
      className={classNames(styles.base, {
        [styles.important]: important,
        [styles.error]: error,
      })}
    >
      <Group spacing={8} align="center">
        {children}
        {loading && (
          <Spinner animation="border" role="status" className={styles.spinner}>
            <span className="sr-only">Loading...</span>
          </Spinner>
        )}
        <SuccessCheckmark
          show={actionable && success}
          className={styles.checkmark}
        />
      </Group>
    </BootstrapModalHeader>
  );
};

export default ModalHeader;
