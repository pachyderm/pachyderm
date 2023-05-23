import {Project} from '@graphqlTypes';
import classNames from 'classnames';
import noop from 'lodash/noop';
import React from 'react';
import {useHistory} from 'react-router';

import ActiveProjectModal from '@dash-frontend/components/ActiveProjectModal';
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
  const browserHistory = useHistory();
  const {
    openModal: openActiveProjectModal,
    closeModal: closeActiveProjectModal,
    isOpen: activeProjectModalIsOpen,
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

  const onClick = () =>
    browserHistory.push(lineageRoute({projectId: project.id}));

  const overflowMenuItems: DropdownItem[] = [
    {
      id: 'set-active',
      content: 'Set Active Project',
      closeOnClick: true,
    },
    {
      id: 'edit-project-info',
      content: 'Edit Project Info',
      closeOnClick: true,
    },
    {
      id: 'delete-project',
      content: 'Delete Project',
      closeOnClick: true,
    },
  ];

  const onSelect = (id: string) => {
    switch (id) {
      case 'set-active':
        openActiveProjectModal();
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
        className={classNames(styles.row, {
          [styles[`${project.status}Selected`]]: isSelected,
          [styles.rowHover]: multiProject && !isSelected,
        })}
        onClick={() => setSelectedProject()}
        role="row"
      >
        <Group vertical spacing={16}>
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
              />
            </Group>
          </Group>
          <Group spacing={64}>
            <Info header="Project Status" headerId="project-status">
              <ProjectStatus
                status={project.status}
                data-testid="ProjectRow__status"
              />
            </Info>
            <Info
              header="Description"
              headerId="project-description"
              className={styles.responsiveHide}
            >
              {project.description || 'N/A'}
            </Info>
          </Group>
        </Group>
      </div>
      {activeProjectModalIsOpen && (
        <ActiveProjectModal
          show={activeProjectModalIsOpen}
          onHide={closeActiveProjectModal}
          projectName={project.id}
        />
      )}
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
    </>
  );
};

export default ProjectRow;
