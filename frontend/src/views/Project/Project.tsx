import {
  LoadingDots,
  Tooltip,
  CheckboxCheckedSVG,
  CheckboxSVG,
  FullscreenSVG,
  FlipSVG,
  Icon,
} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {Helmet} from 'react-helmet';
import {Route} from 'react-router';

import EmptyState from '@dash-frontend/components/EmptyState';
import {
  NO_DAG_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import View from '@dash-frontend/components/View';
import {DagDirection} from '@dash-frontend/lib/types';
import HoveredNodeProvider from '@dash-frontend/providers/HoveredNodeProvider';
import {useWorkspace} from 'hooks/useWorkspace';

import FileBrowser from '../FileBrowser';
import JobLogsViewer from '../LogsViewers/JobLogsViewer/JobLogsViewer';
import PipelineLogsViewer from '../LogsViewers/PipelineLogsViewer';

import DAG from './components/DAG';
import ProjectHeader from './components/ProjectHeader';
import ProjectSidebar from './components/ProjectSidebar';
import RangeSlider from './components/RangeSlider';
import {NODE_HEIGHT, NODE_WIDTH} from './constants/nodeSizes';
import {
  FILE_BROWSER_PATH,
  LOGS_VIEWER_JOB_PATH,
  LOGS_VIEWER_PIPELINE_PATH,
} from './constants/projectPaths';
import {MAX_SCALE_VALUE, useProjectView} from './hooks/useProjectView';
import styles from './Project.module.css';

const MARKERS = [
  {id: 'end-arrow', color: '#747475'},
  {id: 'end-arrow-active', color: '#6FB3C3'},
  {id: 'end-arrow-error', color: '#E02020'},
];

const Project: React.FC = () => {
  const {
    applySliderZoom,
    dags,
    error,
    loading,
    minScale,
    rotateDag,
    dagDirection,
    sliderZoomValue,
    svgSize,
    zoomOut,
    isSidebarOpen,
    sidebarSize,
    skipCenterOnSelect,
    handleChangeCenterOnSelect,
  } = useProjectView(NODE_WIDTH, NODE_HEIGHT);
  const {hasConnectInfo} = useWorkspace();

  const noDags = dags?.length === 0;
  return (
    <>
      <Helmet>
        <title>Project - Pachyderm Console</title>
      </Helmet>
      <ProjectHeader />
      <View className={styles.view}>
        {loading ? (
          <div
            className={classnames(styles.loadingContainer, {
              [styles.isSidebarOpen]: isSidebarOpen,
              [styles[sidebarSize]]: true,
            })}
          >
            <LoadingDots />
          </div>
        ) : (
          <>
            <div className={styles.canvasControls}>
              <RangeSlider
                min={(minScale * 100).toString()}
                max={(MAX_SCALE_VALUE * 100).toString()}
                onChange={(d: React.ChangeEvent<HTMLInputElement>) =>
                  applySliderZoom(d)
                }
                value={sliderZoomValue * 100}
                disabled={noDags}
              />
              <button
                className={styles.controlButton}
                onClick={rotateDag}
                disabled={noDags}
              >
                <Icon>
                  <FlipSVG
                    aria-label={'Rotate Canvas'}
                    className={classnames(styles.rotateSvg, {
                      [styles.flipped]: dagDirection === DagDirection.RIGHT,
                    })}
                  />
                </Icon>
                <label className={styles.controlLabel}>Flip Canvas</label>
              </button>
              <Tooltip
                className={styles.tooltip}
                tooltipKey="zoomOut"
                size="large"
                placement="bottom"
                tooltipText={`Click to reset canvas, or\nuse keyboard shortcut "Shift + 2"`}
              >
                <button
                  className={styles.controlButton}
                  onClick={zoomOut}
                  disabled={noDags}
                >
                  <Icon>
                    <FullscreenSVG aria-label="Reset Canvas" />
                  </Icon>
                  <label className={styles.controlLabel}>Reset Canvas</label>
                </button>
              </Tooltip>
              <Tooltip
                className={styles.tooltip}
                tooltipKey="skipCenter"
                size="large"
                placement="bottom"
                tooltipText={`${
                  !skipCenterOnSelect ? 'Disable' : 'Enable'
                } panning and zooming to a selection`}
              >
                <button
                  className={styles.controlButton}
                  onClick={() => {
                    handleChangeCenterOnSelect(!!skipCenterOnSelect);
                  }}
                  disabled={noDags}
                >
                  <Icon>
                    {!skipCenterOnSelect ? (
                      <CheckboxCheckedSVG
                        aria-label="Uncheck Skip Center"
                        className={styles.svgControl}
                      />
                    ) : (
                      <CheckboxSVG
                        aria-label="Check Skip Center"
                        className={styles.svgControl}
                      />
                    )}
                  </Icon>
                  <label className={styles.controlLabel}>
                    Center Selections
                  </label>
                </button>
              </Tooltip>
              {error && (
                <span className={styles.dagError}>
                  Connection error: data may not be up to date.
                </span>
              )}
            </div>
            {noDags && (
              <EmptyState
                title={LETS_START_TITLE}
                message={NO_DAG_MESSAGE}
                connect={hasConnectInfo}
              />
            )}
            <HoveredNodeProvider>
              <svg
                id="Svg"
                className={styles.base}
                preserveAspectRatio="xMinYMid meet"
                viewBox={`0 0 ${svgSize.width} ${svgSize.height}`}
              >
                <defs>
                  {MARKERS.map((marker) => (
                    <marker
                      key={marker.id}
                      viewBox="0 -5 10 10"
                      refX={0}
                      refY={0}
                      markerWidth={5}
                      markerHeight={5}
                      orient="auto"
                      id={marker.id}
                    >
                      <path d="M0,-5L10,0L0,5" fill={marker.color} />
                    </marker>
                  ))}
                  <filter id="node-dropshadow">
                    <feDropShadow
                      dx="0"
                      dy="1"
                      stdDeviation="2"
                      floodColor="#C6C6C6"
                    />
                  </filter>
                  <filter id="hover-dropshadow" filterUnits="userSpaceOnUse">
                    <feDropShadow
                      dx="1"
                      dy="1"
                      stdDeviation="6"
                      floodColor="#C6C6C6"
                    />
                  </filter>
                </defs>
                <g id="Dags">
                  {dags?.map((dag) => {
                    return (
                      <DAG
                        data={dag}
                        key={dag.id}
                        id={dag.id}
                        nodeWidth={NODE_WIDTH}
                        nodeHeight={NODE_HEIGHT}
                        dagsToShow={dags.length}
                        dagDirection={dagDirection}
                        rotateDag={rotateDag}
                      />
                    );
                  })}
                </g>
              </svg>
            </HoveredNodeProvider>
          </>
        )}
        <ProjectSidebar />
        <Route path={FILE_BROWSER_PATH}>
          <FileBrowser />
        </Route>
        <Route path={LOGS_VIEWER_PIPELINE_PATH}>
          <PipelineLogsViewer />
        </Route>
        <Route path={LOGS_VIEWER_JOB_PATH}>
          <JobLogsViewer />
        </Route>
      </View>
    </>
  );
};

export default Project;
