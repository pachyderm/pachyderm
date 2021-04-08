import {Info, Group, InfoSVG} from '@pachyderm/components';
import React from 'react';
import {Link} from 'react-router-dom';

import JobList from '@dash-frontend/components/JobList';
import {useProjectDetails} from '@dash-frontend/hooks/useProjectDetails';
import {Project} from '@graphqlTypes';

import ProjectStatus from '../ProjectStatus';

import styles from './ProjectPreview.module.css';

type ProjectPreviewProps = {
  project: Project;
};

const ProjectPreview: React.FC<ProjectPreviewProps> = ({project}) => {
  const {projectDetails, loading} = useProjectDetails(project.id);

  if (!projectDetails || loading) return null; // TODO loading view

  return (
    <div className={styles.base}>
      <div className={styles.topContent}>
        <Group spacing={24} vertical>
          <h4 className={styles.title}>Project Preview</h4>
          <Info
            header="Total No. of Repos/Pipelines"
            headerId="total-repos-and-pipelines"
          >
            {projectDetails.repoCount}/{projectDetails.repoCount}
          </Info>
          <Info header="Total Data Size" headerId="total-data-size">
            {Number(projectDetails.sizeGBytes.toFixed(8))}GB
          </Info>
          <Info header="Pipeline Status" headerId="pipeline-status">
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
          </Info>
          <Group.Divider />
          <h4 className={styles.subTitle}>Last 30 Jobs</h4>
        </Group>
      </div>
      <JobList projectId={project.id} />
    </div>
  );
};

export default ProjectPreview;
