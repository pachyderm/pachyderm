import React from 'react';

import {ProjectInfo} from '@dash-frontend/api/pfs';
import EmptyState from '@dash-frontend/components/EmptyState';
import {Group} from '@pachyderm/components';

import ProjectRow from '../ProjectRow';

import styles from './ProjectTab.module.css';

type ProjectTabProps = {
  noProjects: boolean;
  filteredProjects: ProjectInfo[];
  selectedProject?: ProjectInfo;
  setSelectedProject: React.Dispatch<
    React.SetStateAction<ProjectInfo | undefined>
  >;
  setMyProjectsCount: React.Dispatch<React.SetStateAction<number>>;
  setAllProjectsCount: React.Dispatch<React.SetStateAction<number>>;
  showOnlyAccessible?: boolean;
};

const ProjectTab: React.FC<ProjectTabProps> = ({
  noProjects,
  filteredProjects,
  setSelectedProject,
  setMyProjectsCount,
  setAllProjectsCount,
  selectedProject,
  showOnlyAccessible,
}) => {
  return (
    <Group className={styles.projectsList} spacing={16} vertical>
      {noProjects ? (
        <EmptyState
          title="No projects exist."
          message="Create a project to get started."
        />
      ) : filteredProjects.length === 0 ? (
        <EmptyState
          title="No projects match your current filters."
          message="Try adjusting or resetting your filters to see more projects."
        />
      ) : (
        filteredProjects.map((project) => (
          <ProjectRow
            showOnlyAccessible={showOnlyAccessible}
            multiProject={true}
            project={project}
            key={project?.project?.name}
            setSelectedProject={() => setSelectedProject(project)}
            setMyProjectsCount={setMyProjectsCount}
            setAllProjectsCount={setAllProjectsCount}
            isSelected={
              project?.project?.name === selectedProject?.project?.name
            }
          />
        ))
      )}
    </Group>
  );
};

export default ProjectTab;
