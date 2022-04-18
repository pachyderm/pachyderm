import {Link, RepoSVG, PipelineSVG, Icon} from '@pachyderm/components';
import React, {useMemo} from 'react';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

import styles from './RepoPipelineLink.module.css';

type RepoPipelineLinkProps = {
  name: string;
  branch?: string;
  repo?: string;
  type: 'repo' | 'pipeline' | 'commit';
};

const getIcon = (type: RepoPipelineLinkProps['type']) => {
  switch (type) {
    case 'repo':
      return <RepoSVG />;
    case 'pipeline':
      return <PipelineSVG />;
    default:
      return undefined;
  }
};

const RepoPipelineLink: React.FC<RepoPipelineLinkProps> = ({
  name,
  type,
  branch = 'master',
  repo = '',
  ...rest
}) => {
  const {projectId} = useUrlState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const nodeName = type !== 'pipeline' ? name.replace('_repo', '') : name;
  const path = useMemo(() => {
    switch (type) {
      case 'repo':
        return repoRoute({
          branchId: 'master',
          projectId,
          repoId: nodeName,
        });
      case 'pipeline':
        return pipelineRoute({
          projectId,
          pipelineId: nodeName,
        });
      case 'commit':
        return getPathToFileBrowser({
          projectId,
          commitId: nodeName,
          branchId: branch,
          repoId: repo,
        });

      default:
        return '/not-found';
    }
  }, [nodeName, projectId, type, getPathToFileBrowser, branch, repo]);

  const icon = useMemo(() => getIcon(type), [type]);

  return (
    <Link
      data-testid="RepoPipelineLink__item"
      to={path}
      aria-label={nodeName}
      className={styles.base}
    >
      {icon ? (
        <Icon small className={styles.nodeImage}>
          {icon}
        </Icon>
      ) : null}
      <span {...rest} className={styles.nodeName}>
        {nodeName}
      </span>
    </Link>
  );
};

export default RepoPipelineLink;
