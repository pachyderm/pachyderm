import classNames from 'classnames';
import React from 'react';

import {Button} from '../../../Button';
import {CheckmarkSVG, StatusCheckmarkSVG, InfoSVG} from '../../../Svg';

import styles from './TaskCard.module.css';

type TaskCardProps = {
  task: React.ReactNode;
  action?: () => void;
  index: number;
  currentTask: number;
  actionText?: React.ReactNode;
  taskInfoTitle?: string;
  taskInfo?: React.ReactNode;
};

const TaskCard: React.FC<TaskCardProps> = ({
  children,
  index,
  currentTask,
  action,
  actionText,
  task,
  taskInfoTitle,
  taskInfo,
}) => {
  return (
    <div className={styles.taskCard}>
      <div className={styles.taskHeaderWrapper}>
        <div
          className={`${styles.taskHeader} 
        ${styles.taskHeaderWrapperChild}`}
        >
          <h6 className={styles.task}>{`Task ${index + 1}`}</h6>
          {currentTask > index && (
            <StatusCheckmarkSVG
              aria-label={`Task ${index + 1} complete`}
              className={styles.headerComplete}
            />
          )}
        </div>
        <div className={styles.taskHeaderWrapperChild}>{task}</div>
      </div>
      {taskInfo && taskInfoTitle && (
        <div className={styles.taskInfoWrapper}>
          <InfoSVG className={styles.infoSVG} />
          <strong>{taskInfoTitle}</strong>
          <div className={styles.taskInfo}>{taskInfo}</div>
        </div>
      )}
      {children}
      {action && (
        <div
          className={classNames(styles.action, {
            [styles.completed]: currentTask > index,
          })}
        >
          {currentTask > index && (
            <div className={styles.taskCompleted}>
              <div className={styles.svgWrapper}>
                <CheckmarkSVG />
              </div>
              <div className={styles.completedText}>Task Completed!</div>
            </div>
          )}
          {currentTask <= index && (
            <div>
              <Button
                disabled={currentTask < index}
                onClick={action}
                className={styles.button}
              >
                {actionText}
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default TaskCard;
