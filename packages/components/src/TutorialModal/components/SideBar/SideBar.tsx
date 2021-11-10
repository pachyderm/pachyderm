import React from 'react';

import {PureCheckbox} from '../../../Checkbox';
import {Step} from '../../lib/types';
import TaskListItem from '../TaskListItem';

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
            steps[currentStep].sections.map((section, i) => (
              <TaskListItem
                index={i}
                currentTask={currentTask}
                task={section.taskName}
                key={i}
              />
            ))}
          <div className={styles.taskItem}>
            <PureCheckbox
              name="Continue"
              selected={false}
              label={'Continue to the next story'}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default SideBar;
