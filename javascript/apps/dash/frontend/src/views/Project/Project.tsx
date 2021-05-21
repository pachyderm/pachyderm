import classnames from 'classnames';
import React from 'react';

import View from '@dash-frontend/components/View';
import HoveredNodeProvider from '@dash-frontend/providers/HoveredNodeProvider';

import DAG from './components/DAG';
import ProjectHeader from './components/ProjectHeader';
import ProjectSidebar from './components/ProjectSidebar';
import RangeSlider from './components/RangeSlider';
import {ReactComponent as RotateSvg} from './components/Rotate.svg';
import {NODE_HEIGHT, NODE_WIDTH} from './constants/nodeSizes';
import {useProjectView} from './hooks/useProjectView';
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
    rotateDag,
    dagDirection,
    sliderZoomValue,
    svgSize,
  } = useProjectView(NODE_WIDTH, NODE_HEIGHT);

  if (error)
    return (
      <View>
        <h1>{JSON.stringify(error)}</h1>
      </View>
    );
  if (loading || !dags)
    return (
      <View>
        <h1>Loading...</h1>
      </View>
    );

  return (
    <>
      <ProjectHeader />
      <View className={styles.view}>
        <div className={styles.canvasControls}>
          <RangeSlider
            min="60"
            max="150"
            handleChange={(d: React.ChangeEvent<HTMLInputElement>) =>
              applySliderZoom(d)
            }
            value={sliderZoomValue * 100}
          />
          <button
            className={classnames(styles.rotate, [styles[dagDirection]])}
            onClick={rotateDag}
          >
            <RotateSvg aria-label={'Rotate Dag'} className={styles.rotateSvg} />
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
              {dags.map((dag) => {
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
      </View>
    </>
  );
};

export default Project;
