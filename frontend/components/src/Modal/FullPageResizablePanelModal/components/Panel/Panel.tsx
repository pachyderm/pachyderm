import classNames from 'classnames';
import React from 'react';
import {Panel, PanelProps} from 'react-resizable-panels';

import {Button, CloseSVG} from '@pachyderm/components';

import styles from './Panel.module.css';

export interface ModalPanelProps extends PanelProps {
  children?: React.ReactNode;
  headerContent?: React.ReactNode;
  onClose?: () => void;
  showClose?: boolean;
}

const ModalPanel = ({
  headerContent,
  onClose,
  children,
  showClose = false,
  ...props
}: ModalPanelProps) => {
  return (
    <Panel className={styles.panel} {...props}>
      {showClose && (
        <div className={classNames(showClose && styles.rightHeading)}>
          {headerContent}

          <Button
            data-testid="Panel__closeModal"
            aria-label="Close"
            onClick={onClose}
            className={styles.closeButton}
            IconSVG={CloseSVG}
            buttonType="ghost"
            color="purple"
            iconPosition="end"
          >
            Exit
          </Button>
        </div>
      )}
      <div className={styles.children}>{children}</div>
    </Panel>
  );
};

export default ModalPanel;
