import {useOutsideClick} from '@pachyderm/components';
import noop from 'lodash/noop';
import React, {useRef} from 'react';

import styles from './LeaveJobButton.module.css';

interface LeaveJobButtonProps {
  onClick?: React.MouseEventHandler;
  onClose?: () => void;
  isRepo?: boolean;
}

const HEIGHT = 72;
const Y_OFFSET = -95;
const REPO_X_OFFSET = -50;
const REPO_WIDTH = 230;
const PIPELINE_X_OFFSET = -65;
const PIPELINE_WIDTH = 250;

const LeaveJobButton: React.FC<LeaveJobButtonProps> = ({
  onClick = noop,
  onClose = noop,
  isRepo = false,
}) => {
  const buttonRef = useRef<HTMLButtonElement>(null);
  useOutsideClick(buttonRef, onClose);

  return (
    <foreignObject
      x={isRepo ? REPO_X_OFFSET : PIPELINE_X_OFFSET}
      y={Y_OFFSET}
      height={HEIGHT}
      width={isRepo ? REPO_WIDTH : PIPELINE_WIDTH}
      className={styles.base}
    >
      <button ref={buttonRef} className={styles.button} onClick={onClick}>
        Leave Job to See {isRepo ? 'Repo' : 'Pipeline'} Info
      </button>
    </foreignObject>
  );
};

export default LeaveJobButton;
