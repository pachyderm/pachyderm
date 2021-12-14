import classnames from 'classnames';
import React from 'react';

import {Icon} from 'Icon';

import {StatusWarningSVG, StatusCheckmarkSVG, SpinnerSVG} from '../../../Svg';

import styles from './ModalStatus.module.css';

export type ModalStatusProps = {
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
                return <StatusCheckmarkSVG aria-hidden />;
              case 'error':
                return <StatusWarningSVG aria-hidden />;
              case 'updating':
                return <SpinnerSVG aria-hidden />;
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
