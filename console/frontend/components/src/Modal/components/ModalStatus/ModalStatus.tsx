import classnames from 'classnames';
import React from 'react';

import {Icon} from '@pachyderm/components';

import {
  StatusWarningSVG,
  StatusCheckmarkSVG,
  StatusUpdatedSVG,
} from '../../../Svg';

import styles from './ModalStatus.module.css';

export type ModalStatusProps = {
  children?: React.ReactNode;
  status: 'success' | 'error' | 'updating';
};

const ModalStatus: React.FC<ModalStatusProps> = ({status, children}) => {
  return (
    <div
      role="status"
      className={classnames(styles.base, {
        [styles.success]: status === 'success',
        [styles.error]: status === 'error',
      })}
    >
      <div className={styles.content}>
        <Icon className={styles.icon} color="grey">
          {(() => {
            switch (status) {
              case 'success':
                return (
                  <Icon color="green">
                    <StatusCheckmarkSVG aria-hidden />
                  </Icon>
                );
              case 'error':
                return (
                  <Icon color="red">
                    <StatusWarningSVG aria-hidden />
                  </Icon>
                );
              case 'updating':
                return (
                  <Icon color="green">
                    <StatusUpdatedSVG aria-hidden />
                  </Icon>
                );
              default:
                return null;
            }
          })()}
        </Icon>
        {children}
      </div>
    </div>
  );
};

export default ModalStatus;
