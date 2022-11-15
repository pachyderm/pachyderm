import {Project} from '@graphqlTypes';
import classNames from 'classnames';
import {format, fromUnixTime} from 'date-fns';
import noop from 'lodash/noop';
import React, {useCallback} from 'react';
import {useHistory} from 'react-router';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import {
  projectReposRoute,
  lineageRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Button, Group, Info} from '@pachyderm/components';

import ProjectStatus from '../ProjectStatus';

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
  const [listDefaultView] = useLocalProjectSettings({
    projectId: project.id,
    key: 'list_view_default',
  });
  const onClick = useCallback(() => {
    listDefaultView
      ? browserHistory.push(projectReposRoute({projectId: project.id}))
      : browserHistory.push(lineageRoute({projectId: project.id}));
  }, [browserHistory, listDefaultView, project.id]);
  const date = format(fromUnixTime(project.createdAt), 'MM/d/yyyy');

  return (
    <tr
      className={classNames(styles.row, {
        [styles[`${project.status}Selected`]]: isSelected,
        [styles.rowHover]: multiProject && !isSelected,
      })}
      onClick={(e: React.MouseEvent<HTMLTableRowElement, MouseEvent>) =>
        setSelectedProject()
      }
    >
      <td>
        <Group vertical spacing={16}>
          <Group justify="between" align="baseline" spacing={16}>
            <h5>{project.name}</h5>
            <Button
              buttonType="secondary"
              onClick={onClick}
              className={styles.button}
            >
              <span>View</span>
              <span className={styles.responsiveHide}> Project</span>
            </Button>
          </Group>
          <Group spacing={64}>
            <Info header="Project Status" headerId="project-status">
              <ProjectStatus
                status={project.status}
                data-testid="ProjectRow__status"
              />
            </Info>
            <Info header="Created On" headerId="project-creation">
              <span data-testid="ProjectRow__created">{date}</span>
            </Info>
            <Info
              header="Description"
              headerId="project-description"
              className={styles.responsiveHide}
            >
              {project.description}
            </Info>
          </Group>
        </Group>
      </td>
    </tr>
  );
};

export default ProjectRow;
