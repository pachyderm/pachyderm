import classNames from 'classnames';
import React from 'react';

import {Button} from '../../../Button';
import {
  CheckmarkSVG,
  StatusWarningSVG,
  StatusCheckmarkSVG,
  InfoSVG,
} from '../../../Svg';
import {ErrorText} from '../../../Text';

import styles from './TaskCard.module.css';

type TaskCardProps = {
  task: React.ReactNode;
  action?: () => void;
  error?: string;
  index: number;
  currentTask: number;
  actionText?: React.ReactNode;
  taskInfoTitle?: string;
  taskInfo?: React.ReactNode;
  disabled?: boolean;
};

const TaskCard: React.FC<TaskCardProps> = ({
  children,
  index,
  currentTask,
  action,
  actionText,
  error,
  task,
  taskInfoTitle,
  taskInfo,
  disabled = false,
}) => {
  return (
    <div className={styles.taskCard}>
      <div className={styles.taskHeaderWrapper}>
        <div
          className={`${styles.taskHeader}
        ${styles.taskHeaderWrapperChild}`}
        >
          <h5 className={styles.task}>{`Task ${index + 1}`}</h5>
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
          <h6>{taskInfoTitle}</h6>
          <div className={styles.taskInfo}>{taskInfo}</div>
        </div>
      )}
      {children}
      {action && (
        <div
          className={classNames(styles.action, {
            [styles.completed]: currentTask > index,
            [styles.error]: error,
          })}
        >
          {currentTask > index && (
            <div className={styles.buttonCover}>
              <div className={styles.svgWrapper}>
                <CheckmarkSVG />
              </div>
              <strong className={styles.completedText}>Task Completed!</strong>
            </div>
          )}
          {error && (
            <div className={styles.buttonCover}>
              <div className={styles.svgWrapperError}>
                <StatusWarningSVG />
              </div>
              <ErrorText className={styles.errorText}>
                {error} - you may need to restart the tutorial
              </ErrorText>
            </div>
          )}
          {currentTask <= index && !error && (
            <div>
              <Button
                disabled={currentTask < index || disabled}
                onClick={action}
                className={styles.button}
                data-testid={`TaskCard__${actionText}`}
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
