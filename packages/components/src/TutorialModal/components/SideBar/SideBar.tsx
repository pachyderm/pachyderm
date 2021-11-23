import React from 'react';

import {Dropdown, DropdownProps} from 'Dropdown';
import ProgressBar from 'ProgressBar';

import {Story} from '../../lib/types';

import styles from './SideBar.module.css';

const SideBar = ({
  currentStory,
  currentTask,
  handleStoryChange,
  stories,
}: {
  currentStory: number;
  currentTask: number;
  handleStoryChange: DropdownProps['onSelect'];
  stories: Story[];
}) => {
  return (
    <div className={styles.base}>
      <div className={styles.wrapper}>
        <div className={styles.caption}>
          {`Story ${currentStory + 1} of ${stories.length}`}
        </div>
        <Dropdown
          storeSelected
          initialSelectId={stories[0].name}
          className={styles.dropdown}
          onSelect={handleStoryChange}
        >
          <Dropdown.Button className={styles.dropdownButton}>
            {stories[currentStory].name}
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
          {stories[currentStory] &&
            stories[currentStory].sections.map((section, i) => (
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
