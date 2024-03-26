import classNames from 'classnames';
import noop from 'lodash/noop';
import React from 'react';
import {useHistory} from 'react-router';

import ProjectRolesModal from '@dash-frontend/components/ProjectRolesModal';
import {ProjectInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {
  Permission,
  ResourceType,
  useAuthorize,
} from '@dash-frontend/hooks/useAuthorize';
import {useProjectStatus} from '@dash-frontend/hooks/useProjectStatus';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {
  lineageRoute,
  projectConfigRoute,
  projectPipelinesRoute,
  projectReposRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {
  Button,
  DefaultDropdown,
  DropdownItem,
  Group,
  Info,
  LoadingDots,
  OverflowSVG,
  PipelineSVG,
  RepoSVG,
  useModal,
} from '@pachyderm/components';

import DeleteProjectModal from '../DeleteProjectModal';
import ProjectStatus from '../ProjectStatus';
import UpdateProjectModal from '../UpdateProjectModal';

import styles from './ProjectRow.module.css';

type ProjectRowProps = {
  multiProject: boolean;
  project: ProjectInfo;
  isSelected: boolean;
  setSelectedProject: () => void;
  showOnlyAccessible?: boolean;
};

const ProjectRow: React.FC<ProjectRowProps> = ({
  multiProject,
  project,
  isSelected = false,
  setSelectedProject = noop,
  showOnlyAccessible,
}) => {
  const {projectStatus} = useProjectStatus(project.project?.name || '');
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
    isAuthActive,
    hasProjectDelete,
    hasProjectCreate,
    hasProjectModifyBindings,
    hasProjectSetDefaults,
    hasProjectListRepo,
    hasRepoRead,
    loading,
  } = useAuthorize({
    permissions: [
      Permission.PROJECT_DELETE,
      Permission.PROJECT_CREATE,
      Permission.PROJECT_MODIFY_BINDINGS,
      Permission.PROJECT_SET_DEFAULTS,
      Permission.PROJECT_LIST_REPO,
      Permission.REPO_READ,
    ],
    resource: {type: ResourceType.PROJECT, name: project.project?.name || ''},
  });

  const show = showOnlyAccessible ? hasProjectListRepo || hasRepoRead : true;

  const onViewDAGClick = () =>
    browserHistory.push(lineageRoute({projectId: project.project?.name || ''}));

  const onViewReposClick = () =>
    browserHistory.push(
      projectReposRoute({projectId: project.project?.name || ''}),
    );

  const onViewPipelinesClick = () =>
    browserHistory.push(
      projectPipelinesRoute({projectId: project.project?.name || ''}),
    );

  const overflowMenuItems: DropdownItem[] = [];

  if (hasProjectCreate)
    overflowMenuItems.push({
      id: 'edit-project-info',
      content: 'Edit Project Info',
      closeOnClick: true,
    });

  if (isAuthActive)
    overflowMenuItems.push({
      id: 'edit-project-roles',
      content: hasProjectModifyBindings
        ? 'Edit Project Roles'
        : 'View Project Roles',
      closeOnClick: true,
    });

  if (hasProjectSetDefaults)
    overflowMenuItems.push({
      id: 'project-defaults',
      content: 'Edit Project Defaults',
      closeOnClick: true,
    });

  if (hasProjectDelete)
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
      case 'project-defaults':
        browserHistory.push(
          projectConfigRoute({projectId: project.project?.name || ''}),
        );
        return null;
      default:
        return null;
    }
  };

  return (
    <>
      {loading && (
        <div
          className={classNames(styles.row, styles.minHeight10, styles.loading)}
          role="row"
        >
          <LoadingDots />
        </div>
      )}
      {!loading && show && (
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
                <h5>{project.project?.name || ''}</h5>
                <Group spacing={8}>
                  <Button
                    buttonType="secondary"
                    onClick={onViewDAGClick}
                    className={styles.button}
                    aria-label={`View project ${project.project?.name || ''}`}
                  >
                    <span>View</span>
                    <span className={styles.responsiveHide}> Project</span>
                  </Button>
                  <Button
                    buttonType="ghost"
                    onClick={onViewPipelinesClick}
                    className={styles.button}
                    aria-label={`View ${project.project?.name || ''} pipelines`}
                    IconSVG={PipelineSVG}
                  >
                    <span className={styles.responsiveHide}>
                      View Pipelines
                    </span>
                  </Button>
                  <Button
                    buttonType="ghost"
                    onClick={onViewReposClick}
                    className={styles.button}
                    aria-label={`View ${project.project?.name || ''} repos`}
                    IconSVG={RepoSVG}
                  >
                    <span className={styles.responsiveHide}>View Repos</span>
                  </Button>
                  <DefaultDropdown
                    items={overflowMenuItems}
                    onSelect={onSelect}
                    aria-label={`${project.project?.name || ''} overflow menu`}
                    buttonOpts={{
                      hideChevron: true,
                      IconSVG: OverflowSVG,
                      buttonType: 'ghost',
                    }}
                    menuOpts={{pin: 'right'}}
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
                {project.createdAt && (
                  <Info header="Created On" headerId="created-date">
                    <span data-testid="ProjectRow__created">
                      {getStandardDateFromISOString(project.createdAt)}
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
              projectName={project.project?.name || ''}
              description={project.description}
            />
          )}
          {deleteModalIsOpen && (
            <DeleteProjectModal
              show={deleteModalIsOpen}
              onHide={closeDeleteModal}
              projectName={project.project?.name || ''}
            />
          )}
          {rolesModalOpen && (
            <ProjectRolesModal
              show={rolesModalOpen}
              onHide={closeRolesModal}
              projectName={project.project?.name || ''}
              readOnly={!hasProjectModifyBindings}
            />
          )}
        </>
      )}
    </>
  );
};

export default ProjectRow;
