import {Tooltip} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {Route} from 'react-router';

import LoadingSkeleton from '@dash-frontend/components/LoadingSkeleton';
import View from '@dash-frontend/components/View';
import HoveredNodeProvider from '@dash-frontend/providers/HoveredNodeProvider';

import DAG from './components/DAG';
import FileBrowser from './components/FileBrowser';
import ProjectHeader from './components/ProjectHeader';
import ProjectSidebar from './components/ProjectSidebar';
import RangeSlider from './components/RangeSlider';
import {ReactComponent as RotateSvg} from './components/Rotate.svg';
import {ReactComponent as ZoomOutSvg} from './components/ZoomOut.svg';
import {NODE_HEIGHT, NODE_WIDTH} from './constants/nodeSizes';
import {FILE_BROWSER_PATH} from './constants/projectPaths';
import {MAX_SCALE_VALUE, useProjectView} from './hooks/useProjectView';
import styles from './Project.module.css';

const MARKERS = [
  {id: 'end-arrow', color: '#000'},
  {id: 'end-arrow-active', color: '#5ba3b1'},
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
  } = useProjectView(NODE_WIDTH, NODE_HEIGHT);

  if (error)
    return (
      <View>
        <h1>{JSON.stringify(error)}</h1>
      </View>
    );

  if (loading) return <LoadingSkeleton />;

  return (
    <>
      <ProjectHeader />
      <View className={styles.view}>
        <div className={styles.canvasControls}>
          <RangeSlider
            min={(minScale * 100).toString()}
            max={(MAX_SCALE_VALUE * 100).toString()}
            onChange={(d: React.ChangeEvent<HTMLInputElement>) =>
              applySliderZoom(d)
            }
            value={sliderZoomValue * 100}
            disabled={dags?.length === 0}
          />
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
              disabled={dags?.length === 0}
            >
              <ZoomOutSvg
                aria-label="Reset Canvas"
                className={classnames(styles.svgControl, styles.zoomOutSvg)}
              />
              <label className={styles.controlLabel}>Reset Canvas</label>
            </button>
          </Tooltip>
          <button
            className={classnames(styles.controlButton, [styles[dagDirection]])}
            onClick={rotateDag}
            disabled={dags?.length === 0}
          >
            <RotateSvg
              aria-label={'Rotate Canvas'}
              className={classnames(styles.svgControl, styles.rotateSvg)}
            />
          </button>
        </div>
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
                  refX={9}
                  markerWidth={5}
                  markerHeight={5}
                  orient="auto"
                  id={marker.id}
                >
                  <path d="M0,-5L10,0L0,5" fill={marker.color} />
                </marker>
              ))}
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
        <ProjectSidebar />
        <Route path={FILE_BROWSER_PATH}>
          <FileBrowser />
        </Route>
      </View>
    </>
  );
};

export default Project;
