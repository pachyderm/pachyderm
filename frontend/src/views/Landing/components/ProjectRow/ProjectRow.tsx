import {Project} from '@graphqlTypes';
import {Button, Info} from '@pachyderm/components';
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
        [styles.rowHover]: multiProject,
      })}
      onClick={(e: React.MouseEvent<HTMLTableRowElement, MouseEvent>) =>
        setSelectedProject()
      }
    >
      <td>
        <h3 className={styles.title}>{project.name}</h3>
      </td>
      <td>
        <Info header="Project Status" headerId="project-status">
          <ProjectStatus
            status={project.status}
            data-testid="ProjectRow__status"
          />
        </Info>
      </td>
      <td>
        <Info header="Created On" headerId="project-creation">
          <span data-testid="ProjectRow__created">{date}</span>
        </Info>
      </td>
      <td>
        <Info header="Description" headerId="project-description">
          {project.description}
        </Info>
      </td>
      <td>
        <Button
          className={styles.button}
          buttonType="secondary"
          onClick={onClick}
        >
          View Project
        </Button>
      </td>
    </tr>
  );
};

export default ProjectRow;
