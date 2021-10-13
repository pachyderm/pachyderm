import React from 'react';

import {PureCheckbox} from '../../../Checkbox';
import {Step} from '../../lib/types';

import styles from './SideBar.module.css';

const SideBar = ({
  currentStep,
  currentTask,
  steps,
}: {
  currentStep: number;
  currentTask: number;
  steps: Step[];
}) => {
  return (
    <div className={styles.base}>
      <div className={styles.wrapper}>
        <div className={styles.caption}>
          {steps[currentStep].label || `Step ${currentStep}`}
        </div>
        <div className={styles.subHeaderXS}>{steps[currentStep]?.name}</div>
        <div className={styles.taskList}>
          {steps[currentStep] &&
            steps[currentStep].tasks.map((task, i) => (
              <div className={styles.taskItem} key={task.name?.toString()}>
                <PureCheckbox
                  selected={currentTask > i}
                  readOnly
                  label={task.name}
                  checked={currentTask > i}
                />
                {i === currentTask && task.info && (
                  <div className={styles.taskInfo}>
                    <span>{task.info.name}</span>
                    {task.info.text.map((text) => (
                      <p key={text?.toString()}>{text}</p>
                    ))}
                  </div>
                )}
              </div>
            ))}
          <div className={styles.taskItem}>
            <PureCheckbox
              name="Continue"
              selected={false}
              label={'Continue to next step.'}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default SideBar;
