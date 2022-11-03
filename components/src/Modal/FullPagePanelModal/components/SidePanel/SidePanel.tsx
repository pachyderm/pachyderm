import classNames from 'classnames';
import React, {useCallback, useEffect, useRef, useState} from 'react';

import {Button} from 'Button';
import useOutsideClick from 'hooks/useOutsideClick';
import usePanelModal from 'Modal/FullPagePanelModal/hooks/usePanelModal';
import {CloseSVG, PanelLeftSVG, PanelRightSVG} from 'Svg';

import styles from './SidePanel.module.css';

export interface SidePanelProps {
  type: 'left' | 'right';
  open?: boolean;
  isExpanded?: boolean;
  setIsExpanded?: React.Dispatch<React.SetStateAction<boolean>>;
}
const SidePanel: React.FC<SidePanelProps> = ({
  type = 'right',
  open = true,
  isExpanded = false,
  setIsExpanded,
  children,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [isOpen, setIsOpen] = useState(open);
  const {onHide, setLeftOpen} = usePanelModal();

  const isLeft = type === 'left';
  const chevronDirection = isLeft ? isOpen : !isOpen;

  useEffect(() => {
    if (isLeft) {
      setLeftOpen(isOpen);
    }
  }, [isLeft, isOpen, setLeftOpen]);

  const handleClose = () => {
    if (isExpanded && setIsExpanded) {
      setIsExpanded(false);
    } else {
      setIsOpen(!isOpen);
    }
  };

  const handleOutsideClick = useCallback(() => {
    if (isExpanded && setIsExpanded) {
      setIsExpanded(false);
    }
  }, [isExpanded, setIsExpanded]);

  useOutsideClick(containerRef, handleOutsideClick);

  return (
    <div
      ref={containerRef}
      className={classNames(styles.base, {
        [styles.open]: isOpen,
        [styles.left]: isLeft,
        [styles.leftClosed]: !isOpen && isLeft,
        [styles.expanded]: isOpen && isLeft && isExpanded,
      })}
    >
      {!isLeft && (
        <div
          className={classNames(styles.rightHeading, {
            [styles.open]: isOpen,
          })}
        >
          <Button
            data-testid="SidePanel__closeModal"
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

      {isOpen && <div className={styles.children}>{children}</div>}

      <div
        className={classNames(styles.footer, {
          [styles.left]: type === 'left',
          [styles.closed]: !isOpen,
        })}
      >
        <Button
          color="black"
          buttonType="ghost"
          IconSVG={chevronDirection ? PanelLeftSVG : PanelRightSVG}
          onClick={handleClose}
        />
      </div>
    </div>
  );
};

export const LeftPanel: React.FC<{
  open?: boolean;
  isExpanded?: boolean;
  setIsExpanded?: React.Dispatch<React.SetStateAction<boolean>>;
}> = ({open, isExpanded, setIsExpanded, children}) => {
  return (
    <SidePanel
      type="left"
      open={open}
      isExpanded={isExpanded}
      setIsExpanded={setIsExpanded}
    >
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
