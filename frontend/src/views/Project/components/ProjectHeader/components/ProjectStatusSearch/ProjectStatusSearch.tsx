import React, {useCallback, useEffect, useRef, useState} from 'react';

import {ProjectStatus} from '@dash-frontend/api/pfs';
import {useProjectStatus} from '@dash-frontend/hooks/useProjectStatus';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  Button,
  Icon,
  StatusCheckmarkSVG,
  StatusWarningSVG,
  useOutsideClick,
  usePreviousValue,
} from '@pachyderm/components';

import Dropdown from './components/Dropdown/Dropdown';
import styles from './ProjectStatusSearch.module.css';

const ProjectStatusSearch: React.FC = () => {
  const {projectId} = useUrlState();
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const {projectStatus, loading} = useProjectStatus(projectId);

  const handleOutsideClick = useCallback(() => {
    if (isOpen) {
      setIsOpen(false);
    }
  }, [isOpen, setIsOpen]);
  useOutsideClick(containerRef, handleOutsideClick);

  const prevStatus = usePreviousValue(projectStatus);

  useEffect(() => {
    if (
      prevStatus === ProjectStatus.UNHEALTHY &&
      projectStatus === ProjectStatus.HEALTHY
    ) {
      setIsOpen(false);
    }
  }, [prevStatus, projectStatus]);

  if (loading) return <></>;

  return (
    <div className={styles.base} ref={containerRef}>
      {projectStatus === ProjectStatus.HEALTHY ? (
        <div className={styles.healthyContainer}>
          <Icon color="green" className={styles.icon} small>
            <StatusCheckmarkSVG />
          </Icon>
          <div className={styles.hide}>Healthy Project</div>
        </div>
      ) : (
        <>
          <Button
            buttonType="tertiary"
            onClick={() => setIsOpen((current) => !current)}
          >
            <Icon color="red" small>
              <StatusWarningSVG />
            </Icon>
            <div className={styles.hide}>Unhealthy Project</div>
          </Button>
        </>
      )}

      <div>{isOpen && <Dropdown />}</div>
    </div>
  );
};

export default ProjectStatusSearch;
