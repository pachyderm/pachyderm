import React from 'react';
import {useHistory} from 'react-router';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import RepoDetails from '@dash-frontend/views/Project/components/ProjectSidebar/components/RepoDetails';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  CaptionText,
  LoadingDots,
  FullPagePanelModal,
  Button,
  ArrowLeftSVG,
  Pager,
} from '@pachyderm/components';

import FileHeader from './components/FileHeader';
import FileHistory from './components/FileHistory';
import FilePreview from './components/FilePreview';
import LeftPanel from './components/LeftPanel';
import ListViewTable from './components/ListViewTable';
import styles from './FileBrowser.module.css';
import useFileBrowser, {FILE_DEFAULT_PAGE_SIZE} from './hooks/useFileBrowser';

const FileBrowser: React.FC = () => {
  const {repoId, branchId, projectId, filePath, commitId} = useUrlState();
  const browserHistory = useHistory();

  const {
    handleHide,
    isOpen,
    files,
    error,
    loading,
    fileToPreview,
    isDirectory,
    isRoot,
    selectedCommitId,
    handleBackNav,
    pageSize,
    setPageSize,
    page,
    updatePage,
    cursors,
    hasNextPage,
    isCommitOpen,
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
      <BrandedTitle title="Files" />
      <FullPagePanelModal
        show={isOpen}
        onHide={handleHide}
        hideType="exit"
        className={styles.fullModal}
      >
        <LeftPanel selectedCommitId={selectedCommitId} />
        <FullPagePanelModal.Body>
          <div className={styles.base}>
            <FileHeader commitId={selectedCommitId} />
            {!fileToPreview && !loading && (
              <div className={styles.header}>
                <div data-testid="FileBrowser__title">
                  <h6>
                    {isRoot
                      ? 'Commit files for'
                      : `Folder: ${filePath.slice(0, -1)}`}
                  </h6>
                  <CaptionText color="black">{selectedCommitId}</CaptionText>
                </div>
                {!isRoot && (
                  <Button
                    buttonType="secondary"
                    IconSVG={ArrowLeftSVG}
                    onClick={handleBackNav}
                  >
                    Back
                  </Button>
                )}
              </div>
            )}

            {loading && <LoadingDots data-testid="FileBrowser__loadingDots" />}

            {!loading && (
              <>
                {error ? (
                  <ErrorStateSupportLink
                    title="We couldn't load the file list"
                    message="Your files have been processed, but we couldn't fetch a list of them from our end. Please try refreshing this page."
                  />
                ) : (
                  <>
                    {isCommitOpen ? (
                      <EmptyState
                        title="This commit is currently open"
                        message="Your data is being processed, and will be available to preview after the commit has been closed. Please finish the commit to browse these files."
                      />
                    ) : (
                      <>
                        {files?.length === 0 && !fileToPreview && !error && (
                          <ErrorStateSupportLink
                            title="This commit doesn't have any files"
                            message="Some commits don't contain any files. We still wanted to notify you because Pachyderm didn't detect commits on our end."
                          />
                        )}

                        {files?.length > 0 &&
                          (filePath && fileToPreview ? (
                            <FilePreview file={fileToPreview} />
                          ) : (
                            <ListViewTable files={files} />
                          ))}
                        {files?.length > 0 &&
                          !(filePath && fileToPreview) &&
                          (hasNextPage || cursors.length > 1) && (
                            <Pager
                              elementName="file"
                              page={page}
                              updatePage={updatePage}
                              pageSizes={[FILE_DEFAULT_PAGE_SIZE, 25, 50]}
                              nextPageDisabled={!hasNextPage}
                              updatePageSize={setPageSize}
                              pageSize={pageSize}
                              hasTopBorder
                            />
                          )}
                      </>
                    )}
                  </>
                )}
              </>
            )}
          </div>
        </FullPagePanelModal.Body>
        <FullPagePanelModal.RightPanel>
          {!isDirectory ? <FileHistory /> : <RepoDetails />}
        </FullPagePanelModal.RightPanel>
      </FullPagePanelModal>
    </>
  );
};

export default FileBrowser;
