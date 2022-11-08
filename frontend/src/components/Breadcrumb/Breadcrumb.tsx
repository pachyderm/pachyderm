import React, {useMemo} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {Link, HomeSVG, Icon} from '@pachyderm/components';

import styles from './Breadcrumb.module.css';

const BreadCrumb: React.FC = () => {
  const {repoId, commitId, branchId, projectId, filePath} = useUrlState();

  const directories = useMemo(
    () => filePath.split('/').filter((path) => !!path),
    [filePath],
  );

  return (
    <div className={styles.base}>
      <Link
        aria-label="root directory"
        className={styles.link}
        data-testid="Breadcrumb__home"
        to={fileBrowserRoute({
          repoId,
          branchId,
          projectId,
          commitId,
        })}
      >
        <Icon small color="plum">
          <HomeSVG />
        </Icon>
      </Link>
      {directories.map((dir, index) => (
        <span key={dir} className={styles.dir}>
          /{' '}
          <Link
            className={styles.link}
            to={fileBrowserRoute({
              repoId,
              branchId,
              projectId,
              commitId,
              // filePath must end with a slash for folders
              filePath: directories.slice(0, index + 1).join('/') + '/',
            })}
          >
            {dir}
          </Link>
        </span>
      ))}
    </div>
  );
};

export default BreadCrumb;
