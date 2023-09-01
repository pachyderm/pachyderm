import React, {useCallback, useState} from 'react';

import Header from '@dash-frontend/components/Header';
import HeaderDropdown from '@dash-frontend/components/HeaderDropdown';
import {
  SkeletonDisplayText,
  Tooltip,
  Group,
  Button,
  ArrowLeftSVG,
  StatusWarningSVG,
  Icon,
} from '@pachyderm/components';

import GlobalSearch from './components/GlobalSearch';
import useProjectHeader from './hooks/useProjectHeader';
import styles from './ProjectHeader.module.css';

const ProjectHeader = () => {
  const {projectName, loading, error} = useProjectHeader();
  const [showTooltip, setShowTooltip] = useState(false);

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
            tooltipText={projectName}
            disabled={!showTooltip}
            className={styles.projectNameTooltip}
            allowedPlacements={['bottom']}
          >
            <h6 ref={setProjectNameRef} className={styles.projectName}>
              {!error ? (
                projectName
              ) : (
                <span className={styles.errorName}>
                  <Icon small color="white" className={styles.warningIcon}>
                    <StatusWarningSVG />
                  </Icon>
                  {` Project name unknown`}
                </span>
              )}
            </h6>
          </Tooltip>
        )}
      </Group>
      <div className={styles.dividerSearch} />
      <GlobalSearch />
      <HeaderDropdown />
    </Header>
  );
};

export default ProjectHeader;
