import {useOutsideClick} from '@pachyderm/components';
import noop from 'lodash/noop';
import React, {useRef} from 'react';

import {NODE_WIDTH} from '@dash-frontend/views/Project/constants/nodeSizes';

import styles from './LeaveJobButton.module.css';

interface LeaveJobButtonProps {
  onClick?: React.MouseEventHandler;
  onClose?: () => void;
  isRepo?: boolean;
}

const HEIGHT = 72;
const Y_OFFSET = -95;
const BUTTON_WIDTH = 230;

const LeaveJobButton: React.FC<LeaveJobButtonProps> = ({
  onClick = noop,
  onClose = noop,
  isRepo = false,
}) => {
  const buttonRef = useRef<HTMLButtonElement>(null);
  useOutsideClick(buttonRef, onClose);

  return (
    <foreignObject
      x={(NODE_WIDTH - BUTTON_WIDTH) / 2}
      y={Y_OFFSET}
      height={HEIGHT}
      width={BUTTON_WIDTH}
      className={styles.base}
    >
      <button ref={buttonRef} className={styles.button} onClick={onClick}>
        Leave Job to See {isRepo ? 'Repo' : 'Pipeline'} Info
      </button>
    </foreignObject>
  );
};

export default LeaveJobButton;
