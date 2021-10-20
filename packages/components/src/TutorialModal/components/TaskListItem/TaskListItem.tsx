import React from 'react';

import {PureCheckbox} from '../../../Checkbox';
import {Task} from '../../lib/types';

import styles from './TaskListItem.module.css';

type TaskListItemProps = {
  index: number;
  currentTask: number;
  task: Task;
};

const TaskListItem: React.FC<TaskListItemProps> = ({
  index,
  currentTask,
  task,
  children,
}) => {
  return (
    <div className={styles.taskItem}>
      <PureCheckbox
        selected={currentTask > index}
        readOnly
        label={task?.name}
        checked={currentTask > index}
      />
      {children}
    </div>
  );
};

export default TaskListItem;
