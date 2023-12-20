import {Permission, Project, ResourceType} from '@graphqlTypes';
import classNames from 'classnames';
import noop from 'lodash/noop';
import React from 'react';
import {useHistory} from 'react-router';

import ProjectRolesModal from '@dash-frontend/components/ProjectRolesModal';
import {useProjectStatus} from '@dash-frontend/hooks/useProjectStatus';
import {useVerifiedAuthorizationLazy} from '@dash-frontend/hooks/useVerifiedAuthorizationLazy';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  Button,
  DefaultDropdown,
  DropdownItem,
  Group,
  Info,
  OverflowSVG,
  useModal,
} from '@pachyderm/components';

import DeleteProjectModal from '../DeleteProjectModal';
import ProjectStatus from '../ProjectStatus';
import UpdateProjectModal from '../UpdateProjectModal';

import styles from './ProjectRow.module.css';

type ProjectRowProps = {
  multiProject: boolean;
  project: Project;
  isSelected: boolean;
  setSelectedProject: () => void;
};

const ProjectRow: React.FC<ProjectRowProps> = ({
  multiProject,
  project,
  isSelected = false,
  setSelectedProject = noop,
}) => {
  const {projectStatus} = useProjectStatus(project.id);
  const browserHistory = useHistory();
  const {
    openModal: openRolesModal,
    closeModal: closeRolesModal,
    isOpen: rolesModalOpen,
  } = useModal(false);
  const {
    openModal: openUpdateModal,
    closeModal: closeUpdateModal,
    isOpen: updateModalIsOpen,
  } = useModal(false);
  const {
    openModal: openDeleteModal,
    closeModal: closeDeleteModal,
    isOpen: deleteModalIsOpen,
  } = useModal(false);

  const {
    checkRolesPermission: checkRolesPermissionDeleteProject,
    isAuthorizedAction: deleteProjectIsAuthorizedAction,
  } = useVerifiedAuthorizationLazy({
    permissionsList: [Permission.PROJECT_DELETE],
    resource: {type: ResourceType.PROJECT, name: project.id},
  });

  const {
    checkRolesPermission: checkRolesPermissionEditProject,
    isAuthorizedAction: editProjectIsAuthorizedAction,
  } = useVerifiedAuthorizationLazy({
    permissionsList: [Permission.PROJECT_CREATE],
    resource: {type: ResourceType.PROJECT, name: project.id},
  });

  const {
    checkRolesPermission: checkRolesPermissionEditProjectRole,
    isAuthorizedAction: editProjectRoleIsAuthorizedAction,
    isAuthActive,
  } = useVerifiedAuthorizationLazy({
    permissionsList: [Permission.PROJECT_MODIFY_BINDINGS],
    resource: {type: ResourceType.PROJECT, name: project.id},
  });

  const onClick = () =>
    browserHistory.push(lineageRoute({projectId: project.id}));

  const overflowMenuItems: DropdownItem[] = [];

  if (editProjectIsAuthorizedAction)
    overflowMenuItems.push({
      id: 'edit-project-info',
      content: 'Edit Project Info',
      closeOnClick: true,
    });

  if (isAuthActive)
    overflowMenuItems.push({
      id: 'edit-project-roles',
      content: editProjectRoleIsAuthorizedAction
        ? 'Edit Project Roles'
        : 'View Project Roles',
      closeOnClick: true,
    });

  if (deleteProjectIsAuthorizedAction)
    overflowMenuItems.push({
      id: 'delete-project',
      content: 'Delete Project',
      closeOnClick: true,
    });

  const onSelect = (id: string) => {
    switch (id) {
      case 'edit-project-roles':
        openRolesModal();
        return null;
      case 'edit-project-info':
        openUpdateModal();
        return null;
      case 'delete-project':
        openDeleteModal();
        return null;
      default:
        return null;
    }
  };

  return (
    <>
      <div
        className={classNames(styles.row, styles.minHeight10, {
          [styles[`${projectStatus ?? 'HEALTHY'}Selected`]]: isSelected,
          [styles.rowHover]: multiProject && !isSelected,
        })}
        onClick={() => setSelectedProject()}
        role="row"
      >
        <Group vertical className={styles.gap10}>
          <Group justify="between" align="baseline" spacing={16}>
            <h5>{project.id}</h5>
            <Group spacing={8}>
              <Button
                buttonType="secondary"
                onClick={onClick}
                className={styles.button}
                aria-label={`View project ${project.id}`}
              >
                <span>View</span>
                <span className={styles.responsiveHide}> Project</span>
              </Button>
              <DefaultDropdown
                items={overflowMenuItems}
                onSelect={onSelect}
                aria-label={`${project.id} overflow menu`}
                buttonOpts={{
                  hideChevron: true,
                  IconSVG: OverflowSVG,
                  buttonType: 'ghost',
                }}
                menuOpts={{pin: 'right'}}
                openOnClick={() => {
                  checkRolesPermissionDeleteProject();
                  checkRolesPermissionEditProject();
                  checkRolesPermissionEditProjectRole();
                }}
              />
            </Group>
          </Group>
          <Group className={classNames(styles.gap64)}>
            <Info
              header="Project Status"
              headerId="project-status"
              className={styles.noShrink}
            >
              <ProjectStatus
                status={projectStatus}
                data-testid="ProjectRow__status"
              />
            </Info>
            {project.createdAt?.seconds && (
              <Info header="Created On" headerId="created-date">
                <span data-testid="ProjectRow__created">
                  {getStandardDate(project.createdAt?.seconds || 0)}
                </span>
              </Info>
            )}
            <Info
              header="Description"
              headerId="project-description"
              className={styles.responsiveHide}
              wide
            >
              {project.description || 'N/A'}
            </Info>
          </Group>
        </Group>
      </div>
      {updateModalIsOpen && (
        <UpdateProjectModal
          show={updateModalIsOpen}
          onHide={closeUpdateModal}
          projectName={project.id}
          description={project.description}
        />
      )}
      {deleteModalIsOpen && (
        <DeleteProjectModal
          show={deleteModalIsOpen}
          onHide={closeDeleteModal}
          projectName={project.id}
        />
      )}
      {rolesModalOpen && (
        <ProjectRolesModal
          show={rolesModalOpen}
          onHide={closeRolesModal}
          projectName={project.id}
          readOnly={!editProjectRoleIsAuthorizedAction}
        />
      )}
    </>
  );
};

export default ProjectRow;
