import React from 'react';

import {Button} from '../../../Button';
import {PureCheckbox} from '../../../Checkbox';

import styles from './TaskCard.module.css';

type TaskCardProps = {
  name: React.ReactNode;
  action?: () => void;
  index: number;
  currentTask: number;
  actionText?: string;
};

const TaskCard: React.FC<TaskCardProps> = ({
  children,
  index,
  currentTask,
  name,
  action,
  actionText,
}) => {
  return (
    <div className={styles.taskCard}>
      <div className={styles.taskHeaderWrapper}>
        <h6 className={styles.taskHeader}>{`Task ${index + 1}`}</h6>
        {action && (
          <Button autoSize onClick={action} disabled={currentTask !== index}>
            {actionText}
          </Button>
        )}
      </div>
      <PureCheckbox
        selected={currentTask > index}
        label={name}
        readOnly
        checked={currentTask > index}
      />
      {children}
    </div>
  );
};

export default TaskCard;
