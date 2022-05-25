import {
  SkeletonDisplayText,
  Tooltip,
  Group,
  Button,
  ArrowLeftSVG,
} from '@pachyderm/components';
import React, {useCallback, useState} from 'react';

import GlobalFilter from '@dash-frontend/components/GlobalFilter';
import Header from '@dash-frontend/components/Header';
import HeaderButtons from '@dash-frontend/components/HeaderButtons';
import useRunTutorialButton from '@dash-frontend/components/RunTutorialButton/hooks/useRunTutorialButton';
import Search from '@dash-frontend/components/Search';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import useProjectHeader from './hooks/useProjectHeader';
import styles from './ProjectHeader.module.css';

const ProjectHeader = () => {
  const {projectName, loading} = useProjectHeader();
  const {projectId} = useUrlState();
  const [showTooltip, setShowTooltip] = useState(false);
  const {activeTutorial} = useRunTutorialButton(projectId);

  const setProjectNameRef: React.RefCallback<HTMLHeadingElement> = useCallback(
    (element: HTMLHeadingElement | null) => {
      if (element && element.clientWidth < element.scrollWidth) {
        setShowTooltip(true);
      } else setShowTooltip(false);
    },
    [],
  );

  return (
    <Header>
      <Group spacing={16} align="center">
        <Button
          buttonType="tertiary"
          IconSVG={ArrowLeftSVG}
          className={styles.goBack}
          to="/"
          aria-label="Go back to landing page."
        />

        {loading ? (
          <SkeletonDisplayText
            data-testid="ProjectHeader__projectNameLoader"
            className={styles.projectNameLoader}
            color="grey"
          />
        ) : (
          <Tooltip
            tooltipKey="Project Name"
            tooltipText={projectName}
            placement="bottom"
            size="large"
            disabled={!showTooltip}
            className={styles.projectNameTooltip}
          >
            <h6 ref={setProjectNameRef} className={styles.projectName}>
              {projectName}
            </h6>
          </Tooltip>
        )}
      </Group>
      <div className={styles.dividerSearch} />
      <Search />
      <Group align="center">
        <GlobalFilter />
        {!activeTutorial && <div className={styles.divider} />}
        <HeaderButtons projectId={projectId} />
      </Group>
    </Header>
  );
};

export default ProjectHeader;
