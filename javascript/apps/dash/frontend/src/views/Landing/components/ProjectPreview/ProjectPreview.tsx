import {Info, Group, InfoSVG} from '@pachyderm/components';
import React from 'react';
import {Link} from 'react-router-dom';

import {useJobs} from '@dash-frontend/hooks/useJobs';
import {Project} from '@graphqlTypes';

import ProjectStatus from '../ProjectStatus';

import JobListItem from './components/JobListItem';
import styles from './ProjectPreview.module.css';

type ProjectPreviewProps = {
  project: Project;
};

const ProjectRow: React.FC<ProjectPreviewProps> = ({project}) => {
  const {jobs, loading} = useJobs(project.id);

  if (loading) return null; // TODO loading view

  return (
    <div className={styles.base}>
      <div className={styles.topContent}>
        <Group spacing={24} vertical>
          <h4 className={styles.title}>Project Preview</h4>
          <Info
            header="Total No. of Repos/Pipelines"
            headerId="total-repos-and-pipelines"
          >
            {/* TODO: get from API */}
            120 / 56
          </Info>
          <Info header="Total Data Size" headerId="total-data-size">
            {/* TODO: get from API */}
            85GM
          </Info>
          <Info header="Project Status" headerId="pipeline-status">
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
      {jobs.map((job) => (
        <JobListItem job={job} key={job.id} />
      ))}
    </div>
  );
};

export default ProjectRow;
