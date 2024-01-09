import React from 'react';
import {useHistory} from 'react-router-dom';

import ProjectRolesModal from '@dash-frontend/components/ProjectRolesModal';
import {
  Permission,
  ResourceType,
  useAuthorize,
} from '@dash-frontend/hooks/useAuthorize';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  projectReposRoute,
  projectPipelinesRoute,
  lineageRoute,
  projectJobsRoute,
  createPipelineRoute,
  projectConfigRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {
  SideNav,
  DirectionsSVG,
  RepoSVG,
  PipelineSVG,
  JobsSVG,
  AddCircleSVG,
  useModal,
  UserSettingsSVG,
  Dropdown,
  ChevronRightSVG,
  Icon,
  TechCodeSVG,
} from '@pachyderm/components';
import {MEDIUM} from 'constants/breakpoints';

import CreateRepoModal from '../CreateRepoModal';

import styles from './ProjectSidenav.module.css';

const ProjectSideNav: React.FC = () => {
  const browserHistory = useHistory();
  const {
    openModal: openCreateRepoModal,
    closeModal: closeCreateRepoModal,
    isOpen: createRepoModalOpen,
  } = useModal(false);
  const {
    openModal: openUserRolesModal,
    closeModal: closeUserRolesModal,
    isOpen: userRolesModalOpen,
  } = useModal(false);
  const {projectId} = useUrlState();
  const {
    isAuthActive,
    loading: authorizeLoading,
    hasProjectModifyBindings: hasProjectEditRoles,
    hasProjectCreateRepo,
    hasProjectSetDefaults,
  } = useAuthorize({
    permissions: [
      Permission.PROJECT_MODIFY_BINDINGS,
      Permission.PROJECT_CREATE_REPO,
      Permission.PROJECT_SET_DEFAULTS,
    ],
    resource: {type: ResourceType.PROJECT, name: projectId},
  });

  return (
    <div className={styles.base}>
      <SideNav breakpoint={MEDIUM} loading={authorizeLoading}>
        <SideNav.SideNavList>
          <SideNav.SideNavItem
            IconSVG={DirectionsSVG}
            to={lineageRoute({projectId}, false)}
            tooltipContent="DAG"
            showIconWhenExpanded
          >
            DAG
          </SideNav.SideNavItem>
          {hasProjectCreateRepo && (
            <Dropdown sideOpen>
              <Dropdown.Button
                IconSVG={AddCircleSVG}
                iconPosition="start"
                className={styles.createDropdown}
              >
                Create
                <Icon small>
                  <ChevronRightSVG />
                </Icon>
              </Dropdown.Button>
              <Dropdown.Menu className={styles.menu} pin="right">
                <Dropdown.MenuItem
                  closeOnClick
                  onClick={openCreateRepoModal}
                  id="repo"
                >
                  Input Repository
                </Dropdown.MenuItem>
                <Dropdown.MenuItem
                  closeOnClick
                  onClick={() =>
                    browserHistory.push(createPipelineRoute({projectId}))
                  }
                  id="pipeline"
                >
                  Pipeline
                </Dropdown.MenuItem>
              </Dropdown.Menu>
            </Dropdown>
          )}
          {isAuthActive && (
            <SideNav.SideNavItem
              IconSVG={UserSettingsSVG}
              onClick={openUserRolesModal}
              tooltipContent="User Roles"
              className={styles.buttonLink}
              showIconWhenExpanded
            >
              User Roles
            </SideNav.SideNavItem>
          )}
          {hasProjectSetDefaults && (
            <SideNav.SideNavItem
              IconSVG={TechCodeSVG}
              to={projectConfigRoute({projectId}, false)}
              tooltipContent="Project Defaults"
              className={styles.buttonLink}
              showIconWhenExpanded
            >
              Project Defaults
            </SideNav.SideNavItem>
          )}
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
        {createRepoModalOpen && (
          <CreateRepoModal
            show={createRepoModalOpen}
            onHide={closeCreateRepoModal}
          />
        )}
        {userRolesModalOpen && (
          <ProjectRolesModal
            show={userRolesModalOpen}
            onHide={closeUserRolesModal}
            projectName={projectId}
            readOnly={!hasProjectEditRoles}
          />
        )}
      </SideNav>
    </div>
  );
};

export default ProjectSideNav;
