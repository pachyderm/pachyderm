import classNames from 'classnames';
import React, {useState} from 'react';

import {Button} from 'Button';
import usePanelModal from 'Modal/FullPagePanelModal/hooks/usePanelModal';
import {ChevronDoubleLeftSVG, ChevronDoubleRightSVG, CloseSVG} from 'Svg';

import styles from './SidePanel.module.css';

export interface SidePanelProps {
  type: 'left' | 'right';
  open?: boolean;
}
const SidePanel: React.FC<SidePanelProps> = ({
  type = 'right',
  open = true,
  children,
}) => {
  const [isOpen, setIsOpen] = useState(open);
  const {onHide} = usePanelModal();

  const isLeft = type === 'left';
  const chevronDirection = isLeft ? isOpen : !isOpen;

  return (
    <div
      className={classNames(styles.base, {
        [styles.open]: isOpen,
        [styles.left]: isLeft,
        [styles.leftClosed]: !isOpen && isLeft,
        [styles.leftFixed]: isLeft,
      })}
    >
      {!isLeft && (
        <div
          className={classNames(styles.rightHeading, {
            [styles.open]: isOpen,
          })}
        >
          <Button
            aria-label="Close"
            onClick={onHide}
            className={styles.closeButton}
            IconSVG={CloseSVG}
            buttonType={isOpen ? 'ghost' : 'secondary'}
            color="purple"
            iconPosition="end"
          >
            {isOpen && 'Exit'}
          </Button>
        </div>
      )}

      {isOpen && <div className={styles.rightBody}>{children}</div>}

      <div
        className={classNames(styles.footer, {
          [styles.left]: type === 'left',
          [styles.closed]: !isOpen,
        })}
      >
        <Button
          color="black"
          buttonType="ghost"
          IconSVG={
            chevronDirection ? ChevronDoubleLeftSVG : ChevronDoubleRightSVG
          }
          onClick={() => setIsOpen(!isOpen)}
        />
      </div>
    </div>
  );
};

export const LeftPanel: React.FC<{open?: boolean}> = ({open, children}) => {
  return (
    <SidePanel type="left" open={open}>
      {children}
    </SidePanel>
  );
};

export const RightPanel: React.FC<{open?: boolean}> = ({open, children}) => {
  return (
    <SidePanel type="right" open={open}>
      {children}
    </SidePanel>
  );
};

export default SidePanel;
