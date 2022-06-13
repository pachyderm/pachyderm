import {
  SideNav,
  DirectionsSVG,
  RepoSVG,
  PipelineSVG,
  JobsSVG,
  ViewListSVG,
  AddCircleSVG,
  useModal,
} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {Route} from 'react-router';

import Badge from '@dash-frontend/components/Badge';
import useLineageListViewSwap from '@dash-frontend/hooks/useLineageListViewSwap';
import {
  projectReposRoute,
  projectPipelinesRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {MEDIUM} from 'constants/breakpoints';

import {LINEAGE_PATH, PROJECT_PATH} from '../../constants/projectPaths';
import CreateRepoModal from '../CreateRepoModal';

import {useProjectSideNav} from './hooks/useProjectSideNav';
import styles from './ProjectSidenav.module.css';

const ProjectSideNav: React.FC = () => {
  const {openModal, closeModal, isOpen} = useModal(false);
  const {projectId, numOfFailedJobs, handleListDefaultView, getJobsLink} =
    useProjectSideNav();
  const swapPath = useLineageListViewSwap();

  return (
    <div className={styles.base}>
      <SideNav breakpoint={MEDIUM} styleMode="light">
        <SideNav.SideNavList>
          <Route path={PROJECT_PATH}>
            <SideNav.SideNavButton
              IconSVG={DirectionsSVG}
              tooltipContent="Switch View"
              to={swapPath}
              onClick={() => handleListDefaultView(false)}
              className={styles.buttonLink}
            >
              View Lineage
            </SideNav.SideNavButton>
          </Route>
          <Route path={LINEAGE_PATH}>
            <SideNav.SideNavButton
              IconSVG={ViewListSVG}
              tooltipContent="Switch View"
              to={swapPath}
              onClick={() => handleListDefaultView(true)}
              className={styles.buttonLink}
            >
              View List
            </SideNav.SideNavButton>
          </Route>
        </SideNav.SideNavList>
        <SideNav.SideNavList>
          <Route path={PROJECT_PATH}>
            <SideNav.SideNavItem>
              <SideNav.SideNavLink
                IconSVG={RepoSVG}
                to={projectReposRoute({projectId})}
                tooltipContent="Repos"
                styleMode="light"
                showIconWhenExpanded
              >
                Repositories
              </SideNav.SideNavLink>
            </SideNav.SideNavItem>
            <SideNav.SideNavItem>
              <SideNav.SideNavLink
                IconSVG={PipelineSVG}
                to={projectPipelinesRoute({projectId})}
                tooltipContent="Pipelines"
                styleMode="light"
                showIconWhenExpanded
              >
                Pipelines
              </SideNav.SideNavLink>
            </SideNav.SideNavItem>
          </Route>
          <SideNav.SideNavItem>
            <SideNav.SideNavButton
              IconSVG={AddCircleSVG}
              onClick={openModal}
              tooltipContent="Create New Repo"
              className={classnames(styles.buttonLink, styles.light)}
              buttonType="secondary"
              styleMode="light"
            >
              Create Repo
            </SideNav.SideNavButton>
          </SideNav.SideNavItem>
        </SideNav.SideNavList>
        <SideNav.SideNavList>
          <SideNav.SideNavItem>
            <SideNav.SideNavLink
              IconSVG={JobsSVG}
              tooltipContent="Jobs"
              styleMode="light"
              showIconWhenExpanded
              to={getJobsLink()}
              data-testid="ProjectSideNav__seeJobs"
            >
              {numOfFailedJobs > 0 && (
                <Badge
                  className={styles.seeJobsBadge}
                  aria-label="Number of failed jobs"
                  data-testid="ProjectSideNav__seeJobsBadge"
                >
                  {numOfFailedJobs}
                </Badge>
              )}
              <span data-testid="ProjectSideNav__seeJobsLabel">Jobs</span>
            </SideNav.SideNavLink>
          </SideNav.SideNavItem>
        </SideNav.SideNavList>
        <CreateRepoModal show={isOpen} onHide={closeModal} />
      </SideNav>
    </div>
  );
};

export default ProjectSideNav;
