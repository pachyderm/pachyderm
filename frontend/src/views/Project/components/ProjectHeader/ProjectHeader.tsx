import {SkeletonDisplayText, Tooltip} from '@pachyderm/components';
import React, {useCallback, useState} from 'react';
import {Link} from 'react-router-dom';

import GlobalFilter from '@dash-frontend/components/GlobalFilter';
import Header from '@dash-frontend/components/Header';
import Search from '@dash-frontend/components/Search';

import {ReactComponent as BackArrowSvg} from './BackArrow.svg';
import useProjectHeader from './hooks/useProjectHeader';
import styles from './ProjectHeader.module.css';

const ProjectHeader = () => {
  const {projectName, loading} = useProjectHeader();
  const [showTooltip, setShowTooltip] = useState(false);
  const setProjectNameRef = useCallback((element: HTMLHeadingElement) => {
    setShowTooltip(element && element.clientWidth < element.scrollWidth);
  }, []);

  return (
    <Header>
      <Link to="/" className={styles.goBack}>
        <BackArrowSvg
          aria-label="Go back to landing page."
          className={styles.goBackSvg}
        />
      </Link>

      <div className={styles.projectNameWrapper}>
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
            <h1 ref={setProjectNameRef} className={styles.projectName}>
              {projectName}
            </h1>
          </Tooltip>
        )}
      </div>
      <Search />
      <GlobalFilter />
    </Header>
  );
};

export default ProjectHeader;
