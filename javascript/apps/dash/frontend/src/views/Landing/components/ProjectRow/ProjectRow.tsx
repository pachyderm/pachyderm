import {Button, Info} from '@pachyderm/components';
import {format, fromUnixTime} from 'date-fns';
import React from 'react';
import {useHistory} from 'react-router';

import {Project} from '@graphqlTypes';

import ProjectStatus from './components/ProjectStatus/ProjectStatus';
import styles from './ProjectRow.module.css';

type ProjectRowProps = {
  project: Project;
};

const ProjectRow: React.FC<ProjectRowProps> = ({project}) => {
  const browserHistory = useHistory();
  const onClick = () => browserHistory.push(`/project/${project.id}`);
  const date = format(fromUnixTime(project.createdAt), 'MM/d/yyyy');

  return (
    <tr className={styles.row}>
      <td>
        <h3 className={styles.title}>{project.name}</h3>
      </td>
      <td>
        <Info header="Project Status" headerId="project-status">
          <ProjectStatus status={project.status} />
        </Info>
      </td>
      <td>
        <Info header="Created On" headerId="project-creation">
          {date}
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
