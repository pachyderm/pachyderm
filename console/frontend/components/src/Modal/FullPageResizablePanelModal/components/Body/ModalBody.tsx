import classnames from 'classnames';
import React from 'react';
import {Panel, PanelProps} from 'react-resizable-panels';

import styles from './ModalBody.module.css';

export type ModalBodyProps = PanelProps;

const ModalBody = ({children, className, ...props}: ModalBodyProps) => {
  return (
    <Panel {...props}>
      <div className={classnames(styles.base, className)}>{children}</div>
    </Panel>
  );
};

export default ModalBody;
