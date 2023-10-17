import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';
import {Route} from 'react-router';

import {DEFAULT_SIDEBAR_SIZE} from '@dash-frontend/hooks/useSidebarInfo';
import DeletePipelineButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/DeletePipelineButton';
import DeleteRepoButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/DeleteRepoButton';
import InspectCommitsButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/InspectCommitsButton';
import ReadLogsButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/ReadLogsButton';
import UpdatePipelineButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/UpdatePipelineButton';
import UploadFilesButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/UploadFilesButton';
import {
  PROJECT_PATH,
  LINEAGE_PATH,
  LINEAGE_REPO_PATH,
  LINEAGE_PIPELINE_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {Button, ButtonGroup, CloseSVG} from '@pachyderm/components';

import useSidebar from './hooks/useSidebar';
import styles from './Sidebar.module.css';
interface SidebarProps extends HTMLAttributes<HTMLDivElement> {
  overlay?: boolean;
  fixed?: boolean;
  onClose?: () => void;
  resizable?: boolean;
}

const Sidebar: React.FC<SidebarProps> = ({
  children,
  overlay = false,
  onClose,
  className,
  fixed = false,
  resizable = false,
  ...rest
}) => {
  const {
    isOpen,
    sidebarWidth,
    dragging,
    setDragging,
    throttleMouseEvent,
    applyMousePosition,
    onDragEnd,
  } = useSidebar();

  return (
    <section
      aria-label="project-sidebar"
      onMouseMove={throttleMouseEvent(applyMousePosition, 50)}
      onMouseUp={onDragEnd}
      onMouseLeave={onDragEnd}
      className={classnames(styles.container, {
        [styles.resizable]: resizable,
        [styles.active]: dragging,
      })}
    >
      {overlay && (
        <div
          className={classnames(styles.overlay, {
            [styles.open]: isOpen,
          })}
        />
      )}

      <div
        className={classnames(
          styles.base,
          {
            [styles.fixed]: fixed,
            [styles.open]: isOpen,
          },
          className,
        )}
        style={{width: resizable ? sidebarWidth : DEFAULT_SIDEBAR_SIZE}}
        {...rest}
      >
        {resizable && (
          <div
            className={classnames(styles.dragBar, {[styles.active]: dragging})}
            onMouseDown={() => setDragging(true)}
          />
        )}
        <Route path={[PROJECT_PATH, LINEAGE_PATH]}>
          <div className={styles.sideBarToolbar}>
            <ButtonGroup>
              <Route path={LINEAGE_PIPELINE_PATH}>
                <ReadLogsButton />
              </Route>
              <Route path={LINEAGE_REPO_PATH}>
                <InspectCommitsButton />
              </Route>
            </ButtonGroup>
            <ButtonGroup>
              <Route path={LINEAGE_REPO_PATH}>
                <UploadFilesButton />
                <DeleteRepoButton />
              </Route>
              <Route path={LINEAGE_PIPELINE_PATH}>
                <UpdatePipelineButton />
                <DeletePipelineButton />
              </Route>
              {onClose && (
                <Button
                  IconSVG={CloseSVG}
                  onClick={onClose}
                  aria-label="Close sidebar"
                  color="black"
                  buttonType="ghost"
                />
              )}
            </ButtonGroup>
          </div>
        </Route>

        {children}
      </div>
    </section>
  );
};

export default Sidebar;
