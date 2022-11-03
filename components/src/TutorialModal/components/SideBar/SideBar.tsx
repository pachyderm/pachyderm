import React from 'react';

import {Dropdown, DropdownProps} from '@pachyderm/components';
import ProgressBar from '@pachyderm/components/ProgressBar';

import {Story, Section} from '../../lib/types';

import styles from './SideBar.module.css';

const SideBar = ({
  currentStory,
  currentTask,
  handleStoryChange,
  stories,
  taskSections,
}: {
  currentStory: number;
  currentTask: number;
  handleStoryChange: DropdownProps['onSelect'];
  stories: Story[];
  taskSections: Section[];
}) => {
  return (
    <div className={styles.base}>
      <div className={styles.wrapper}>
        <Dropdown
          selectedId={stories[currentStory].name}
          initialSelectId={stories[0].name}
          className={styles.dropdown}
          onSelect={handleStoryChange}
        >
          <Dropdown.Button className={styles.dropdownButton}>
            <strong>{stories[currentStory].name}</strong>
          </Dropdown.Button>
          <Dropdown.Menu className={styles.dropdownMenu}>
            {stories.map((story) => {
              return (
                <Dropdown.MenuItem
                  key={story.name}
                  id={story.name}
                  closeOnClick
                >
                  {story.name}
                </Dropdown.MenuItem>
              );
            })}
          </Dropdown.Menu>
        </Dropdown>
        <ProgressBar.Container>
          {taskSections.map((section, i) => (
            <ProgressBar.Step
              key={i}
              id={i.toString()}
              nextStepID={
                i < stories[currentStory].sections.length - 1
                  ? (i + 1).toString()
                  : undefined
              }
            >
              <span className={styles.taskInfo}>{section.taskName}</span>
            </ProgressBar.Step>
          ))}
        </ProgressBar.Container>
      </div>
    </div>
  );
};

export default SideBar;
