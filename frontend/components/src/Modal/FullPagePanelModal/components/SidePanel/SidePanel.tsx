import classNames from 'classnames';
import React, {useCallback, useEffect, useRef, useState} from 'react';

import {
  Button,
  CloseSVG,
  PanelLeftSVG,
  PanelRightSVG,
} from '@pachyderm/components';
import useOutsideClick from '@pachyderm/components/hooks/useOutsideClick';
import usePanelModal from '@pachyderm/components/Modal/FullPagePanelModal/hooks/usePanelModal';

import styles from './SidePanel.module.css';

export interface SidePanelProps {
  children?: React.ReactNode;
  headerContent?: React.ReactNode;
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
  headerContent,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [isOpen, setIsOpen] = useState(open);
  const {onHide, setLeftOpen, setRightOpen} = usePanelModal();

  const isLeft = type === 'left';
  const isRight = type === 'right';
  const chevronDirection = isLeft ? isOpen : !isOpen;

  useEffect(() => {
    if (isLeft) {
      setLeftOpen(isOpen);
    }
    if (isRight) {
      setRightOpen(isOpen);
    }
  }, [isLeft, isOpen, isRight, setLeftOpen, setRightOpen]);

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
      data-testid={isLeft ? 'SidePanel__left' : 'SidePanel__right'}
    >
      {!isLeft && (
        <div
          className={classNames(styles.rightHeading, {
            [styles.open]: isOpen,
          })}
        >
          {headerContent}
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
          data-testid={`SidePanel__close${isLeft ? 'Left' : 'Right'}`}
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
  children?: React.ReactNode;
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

export const RightPanel: React.FC<{
  children?: React.ReactNode;
  headerContent?: React.ReactNode;
  open?: boolean;
}> = ({open, children, headerContent}) => {
  return (
    <SidePanel type="right" open={open} headerContent={headerContent}>
      {children}
    </SidePanel>
  );
};

export default SidePanel;
