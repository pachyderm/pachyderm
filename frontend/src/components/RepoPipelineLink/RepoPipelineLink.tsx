import {Link, RepoSVG, PipelineSVG, Icon} from '@pachyderm/components';
import React, {useMemo} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

import styles from './RepoPipelineLink.module.css';

type RepoPipelineLinkProps = {
  name: string;
  type: 'repo' | 'pipeline';
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

const RepoPipelineLink: React.FC<RepoPipelineLinkProps> = ({name, type}) => {
  const {projectId} = useUrlState();
  const nodeName = type !== 'pipeline' ? name.replace('_repo', '') : name;
  const path = useMemo(() => {
    return type === 'repo'
      ? repoRoute({
          branchId: 'master',
          projectId,
          repoId: nodeName,
        })
      : pipelineRoute({
          projectId,
          pipelineId: nodeName,
        });
  }, [nodeName, projectId, type]);

  return (
    <Link
      data-testid="RepoPipelineLink__item"
      to={path}
      aria-label={nodeName}
      className={styles.base}
    >
      <Icon small className={styles.nodeImage}>
        {getIcon(type)}
      </Icon>
      <span className={styles.nodeName}>{nodeName}</span>
    </Link>
  );
};

export default RepoPipelineLink;
