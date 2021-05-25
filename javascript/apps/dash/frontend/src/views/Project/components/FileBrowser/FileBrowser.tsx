import {
  Tooltip,
  GenericErrorSVG,
  ButtonLink,
  FullPageModal,
  useModal,
  ListViewSVG,
  IconViewSVG,
} from '@pachyderm/components';
import React, {useState, useMemo} from 'react';
import {useHistory} from 'react-router';

import Breadcrumb from '@dash-frontend/components/Breadcrumb';
import {useFiles} from '@dash-frontend/hooks/useFiles';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';

import FileHeader from './components/FileHeader';
import styles from './FileBrowser.module.css';

const FileBrowser: React.FC = () => {
  const {repoId, commitId, branchId, projectId, filePath} = useUrlState();
  const browserHistory = useHistory();
  const [fileFilter, setFileFilter] = useState('');
  const [fileView, setFileView] = useState('icon');
  const {closeModal, isOpen} = useModal(true);
  const {files, loading} = useFiles({
    commitId,
    path: filePath || '/',
    branchName: branchId,
    repoName: repoId,
  });

  const filteredFiles = useMemo(
    () =>
      files.filter(
        (file) =>
          !fileFilter ||
          file.path.toLowerCase().includes(fileFilter.toLowerCase()),
      ),
    [fileFilter, files],
  );

  const fileAtLocation = useMemo(() => {
    const hasFileType = filePath.slice(filePath.lastIndexOf('.') + 1);
    return hasFileType && files.find((file) => file.path === filePath);
  }, [filePath, files]);

  return (
    <FullPageModal
      show={isOpen}
      onHide={() => {
        closeModal();
        setTimeout(
          () => browserHistory.push(repoRoute({projectId, repoId, branchId})),
          500,
        );
      }}
      hideType="exit"
    >
      <div className={styles.base}>
        <FileHeader fileFilter={fileFilter} setFileFilter={setFileFilter} />
        <div className={styles.subHeader}>
          <Breadcrumb />
          {!fileAtLocation && (
            <Tooltip
              tooltipText="Icon View"
              placement="left"
              tooltipKey="Icon View"
            >
              <ButtonLink
                onClick={() => setFileView('icon')}
                disabled={fileView === 'icon'}
                aria-label="switch to icon view"
              >
                <IconViewSVG />
              </ButtonLink>
            </Tooltip>
          )}
          {!fileAtLocation && (
            <Tooltip
              tooltipText="List View"
              placement="left"
              tooltipKey="List View"
            >
              <ButtonLink
                onClick={() => setFileView('list')}
                disabled={fileView === 'list'}
                aria-label="switch to list view"
              >
                <ListViewSVG />
              </ButtonLink>
            </Tooltip>
          )}
        </div>
        {!loading && filteredFiles?.length === 0 && (
          <div className={styles.emptyResults}>
            <GenericErrorSVG />
            <h4 className={styles.emptyHeading}>No Matching Results Found.</h4>
            <p className={styles.emptySubheading}>
              {`The folder or file might have been deleted or doesn't exist.
                  Please try searching another keyword.`}
            </p>
          </div>
        )}
      </div>
    </FullPageModal>
  );
};

export default FileBrowser;
