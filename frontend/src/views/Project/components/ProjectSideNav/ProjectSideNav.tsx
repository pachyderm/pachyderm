import React from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  projectReposRoute,
  projectPipelinesRoute,
  lineageRoute,
  projectJobsRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {
  SideNav,
  DirectionsSVG,
  RepoSVG,
  PipelineSVG,
  JobsSVG,
  AddCircleSVG,
  useModal,
} from '@pachyderm/components';
import {MEDIUM} from 'constants/breakpoints';

import CreateRepoModal from '../CreateRepoModal';

import styles from './ProjectSidenav.module.css';

const ProjectSideNav: React.FC = () => {
  const {openModal, closeModal, isOpen} = useModal(false);
  const {projectId} = useUrlState();

  return (
    <div className={styles.base}>
      <SideNav breakpoint={MEDIUM}>
        <SideNav.SideNavList>
          <SideNav.SideNavItem
            IconSVG={DirectionsSVG}
            to={lineageRoute({projectId}, false)}
            tooltipContent="DAG"
            showIconWhenExpanded
          >
            DAG
          </SideNav.SideNavItem>
          <SideNav.SideNavItem
            IconSVG={AddCircleSVG}
            onClick={openModal}
            tooltipContent="Create New Repo"
            className={styles.buttonLink}
            showIconWhenExpanded
          >
            Create Repo
          </SideNav.SideNavItem>
        </SideNav.SideNavList>
        <SideNav.SideNavList label="Lists">
          <SideNav.SideNavItem
            IconSVG={JobsSVG}
            to={projectJobsRoute({projectId}, false)}
            tooltipContent="Jobs"
            showIconWhenExpanded
          >
            Jobs
          </SideNav.SideNavItem>
          <SideNav.SideNavItem
            IconSVG={PipelineSVG}
            to={projectPipelinesRoute({projectId}, false)}
            tooltipContent="Pipelines"
            showIconWhenExpanded
          >
            Pipelines
          </SideNav.SideNavItem>
          <SideNav.SideNavItem
            IconSVG={RepoSVG}
            to={projectReposRoute({projectId}, false)}
            tooltipContent="Repositories"
            showIconWhenExpanded
          >
            Repositories
          </SideNav.SideNavItem>
        </SideNav.SideNavList>
        <CreateRepoModal show={isOpen} onHide={closeModal} />
      </SideNav>
    </div>
  );
};

export default ProjectSideNav;
