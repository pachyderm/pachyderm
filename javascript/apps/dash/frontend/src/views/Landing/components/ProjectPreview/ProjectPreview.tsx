import {Group, InfoSVG} from '@pachyderm/components';
import React from 'react';
import {Link} from 'react-router-dom';

import Description from '@dash-frontend/components/Description';
import JobListStatic from '@dash-frontend/components/JobList/components/JobListStatic';
import {useProjectDetails} from '@dash-frontend/hooks/useProjectDetails';
import {Project} from '@graphqlTypes';

import ProjectStatus from '../ProjectStatus';

import styles from './ProjectPreview.module.css';

type ProjectPreviewProps = {
  project: Project;
};

const ProjectPreview: React.FC<ProjectPreviewProps> = ({project}) => {
  const {projectDetails, loading} = useProjectDetails(project.id);

  return (
    <div className={styles.base}>
      <div className={styles.topContent}>
        <Group spacing={24} vertical>
          <h4 className={styles.title}>Project Preview</h4>
          <Description term="Total No. of Repos/Pipelines" loading={loading}>
            {projectDetails?.repoCount}/{projectDetails?.repoCount}
          </Description>
          <Description term="Total Data Size" loading={loading}>
            {projectDetails?.sizeDisplay}
          </Description>
          <Description term="Pipeline Status" loading={loading}>
            <div className={styles.inline}>
              <ProjectStatus status={project.status} />
              {project.status.toString() === 'UNHEALTHY' && (
                <span className={styles.extras}>
                  {/* TODO: add tooltip for unhealthy project details */}
                  <InfoSVG />
                  {/* TODO: Link to jobs page */}
                  <Link to={'/'} className={styles.link}>
                    Inspect
                  </Link>
                </span>
              )}
            </div>
          </Description>
          <Group.Divider />
          <h4 className={styles.subTitle}>Last 30 Jobs</h4>
        </Group>
      </div>
      <JobListStatic
        projectId={project.id}
        jobs={projectDetails?.jobs}
        loading={loading}
      />
    </div>
  );
};

export default ProjectPreview;
