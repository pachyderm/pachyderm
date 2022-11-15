import classnames from 'classnames';
import React from 'react';
import {Helmet} from 'react-helmet';
import {Switch, useHistory} from 'react-router';

import Breadcrumb from '@dash-frontend/components/Breadcrumb';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  Tooltip,
  GenericErrorSVG,
  FullPageModal,
  ViewIconSVG,
  ViewListSVG,
  LoadingDots,
  Button,
} from '@pachyderm/components';

import FileHeader from './components/FileHeader';
import FilePreview from './components/FilePreview';
import IconView from './components/IconView';
import ListViewTable from './components/ListViewTable';
import styles from './FileBrowser.module.css';
import useFileBrowser from './hooks/useFileBrowser';

const FileBrowser: React.FC = () => {
  const {repoId, branchId, projectId, filePath, commitId} = useUrlState();
  const browserHistory = useHistory();

  const {
    fileFilter,
    setFileFilter,
    fileView,
    setFileView,
    handleHide,
    isOpen,
    filteredFiles,
    loading,
    fileToPreview,
    isDirectory,
    diffOnly,
    setDiffOnly,
  } = useFileBrowser();

  if (filePath && !loading && !fileToPreview && !isDirectory) {
    browserHistory.push(
      fileBrowserRoute({
        repoId,
        branchId,
        projectId,
        commitId,
      }),
    );
  }

  return (
    <>
      <Helmet>
        <title>Files - Pachyderm Console</title>
      </Helmet>
      <FullPageModal
        show={isOpen}
        onHide={handleHide}
        hideType="exit"
        className={styles.fullModal}
      >
        <div className={styles.base}>
          <FileHeader
            fileFilter={fileFilter}
            setFileFilter={setFileFilter}
            setDiffOnly={setDiffOnly}
            diffOnly={diffOnly}
          />
          <div className={styles.subHeader}>
            <Breadcrumb />
            {isDirectory && (
              <>
                <Tooltip
                  tooltipText="List View"
                  placement="left"
                  tooltipKey="List View"
                >
                  <Button
                    buttonType="ghost"
                    onClick={() => setFileView('list')}
                    className={classnames({
                      [styles.inactiveIcon]: fileView === 'list',
                    })}
                    aria-label="switch to list view"
                    IconSVG={ViewListSVG}
                  />
                </Tooltip>
                <Tooltip
                  tooltipText="Icon View"
                  placement="left"
                  tooltipKey="Icon View"
                >
                  <Button
                    buttonType="ghost"
                    onClick={() => setFileView('icon')}
                    className={classnames({
                      [styles.inactiveIcon]: fileView === 'icon',
                    })}
                    aria-label="switch to icon view"
                    IconSVG={ViewIconSVG}
                  />
                </Tooltip>
              </>
            )}
          </div>
          <Switch>
            {isDirectory ? (
              <>
                {fileView === 'icon' && filteredFiles.length > 0 && (
                  <div
                    className={styles.fileIcons}
                    data-testid="FileBrowser__iconView"
                  >
                    {filteredFiles.map((file) => (
                      <IconView key={file.path} file={file} />
                    ))}
                  </div>
                )}
                {fileView === 'list' && filteredFiles.length > 0 && (
                  <ListViewTable files={filteredFiles} />
                )}
              </>
            ) : (
              <>{fileToPreview && <FilePreview file={fileToPreview} />}</>
            )}
          </Switch>
          {loading && <LoadingDots />}
          {!loading && filteredFiles?.length === 0 && !fileToPreview && (
            <div className={styles.emptyResults}>
              <GenericErrorSVG />
              <h4 className={styles.emptyHeading}>
                No Matching Results Found.
              </h4>
              <p className={styles.emptySubheading}>
                {`The folder or file might have been deleted or doesn't exist.
                  Please try searching another keyword.`}
              </p>
            </div>
          )}
        </div>
      </FullPageModal>
    </>
  );
};

export default FileBrowser;
