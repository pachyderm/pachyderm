import classNames from 'classnames';
import React, {useCallback, useState} from 'react';

import {PureCheckbox} from '../Checkbox';

import SideBar from './components/SideBar';
import TaskListItem from './components/TaskListItem';
import {Step} from './lib/types';
import styles from './TutorialModal.module.css';

type TutorialModalProps = {
  steps: Step[];
};

const TutorialModal: React.FC<TutorialModalProps> = ({steps}) => {
  const [minimized, setMinimized] = useState(false);
  const [currentTask, setCurrentTask] = useState(0);
  const [currentStep, setCurrentStep] = useState(0);

  const classes = classNames(styles.modal, {
    [styles.minimize]: minimized,
    [styles.maximize]: !minimized,
  });

  const handleTaskCompletion = useCallback(
    (index: number) => {
      if (currentTask === index) setCurrentTask((prevValue) => prevValue + 1);
    },
    [currentTask],
  );

  const displayTaskIndex = Math.min(
    currentTask,
    steps[currentStep].tasks.length - 1,
  );
  const currentTaskInstance = steps[currentStep].tasks[displayTaskIndex];

  return (
    <>
      <div className={!minimized ? styles.overlay : ''} />
      <div className={classes}>
        {currentTaskInstance ? (
          <div
            className={classNames(styles.miniTask, {
              [styles.miniTaskOpen]: minimized,
            })}
          >
            <TaskListItem
              currentTask={currentTask}
              index={displayTaskIndex}
              task={currentTaskInstance}
            />
            {displayTaskIndex !== currentTask ? (
              <PureCheckbox
                name="Continue"
                selected={false}
                label={'Continue to next step.'}
              />
            ) : null}
          </div>
        ) : null}
        <div className={styles.header}>
          <button
            className={styles.button}
            onClick={() => setCurrentStep((prevValue) => prevValue + 1)}
            disabled={
              currentStep === steps.length - 1 ||
              currentTask <= steps[currentStep]?.tasks.length - 1
            }
          >
            Continue to next step
          </button>
          <button
            className={styles.button}
            onClick={() => setMinimized((prevValue) => !prevValue)}
          >
            {minimized ? 'Maximize' : 'Minimize'}
          </button>
        </div>
        <div className={styles.body}>
          <SideBar
            currentStep={currentStep}
            currentTask={currentTask}
            steps={steps}
          />
          <div className={styles.content}>
            <div className={styles.instructions}>
              <h2 className={styles.instructionsHeader}>
                {steps[currentStep].instructionsHeader}
              </h2>
              <div className={styles.instructionsBody}>
                <div className={styles.instructionsText}>
                  {steps[currentStep].instructionsText}
                </div>
              </div>
            </div>
            {steps[currentStep].tasks.map((task, i) => {
              return (
                <React.Fragment key={task.name?.toString()}>
                  <task.Task
                    currentTask={currentTask}
                    onCompleted={() => handleTaskCompletion(i)}
                    minimized={minimized}
                    index={i}
                    name={task.name}
                  />
                  <div className={styles.followUp}>{task.followUp}</div>
                </React.Fragment>
              );
            })}
          </div>
        </div>
      </div>
    </>
  );
};

export default TutorialModal;
