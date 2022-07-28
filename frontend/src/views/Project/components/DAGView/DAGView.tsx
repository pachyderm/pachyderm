import {ApolloError} from '@apollo/client';
import {
  Tooltip,
  CheckboxCheckedSVG,
  CheckboxSVG,
  FullscreenSVG,
  FlipSVG,
  Button,
  ButtonGroup,
} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import {
  NO_DAG_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import View from '@dash-frontend/components/View';
import {DagDirection, Dag} from '@dash-frontend/lib/types';
import HoveredNodeProvider from '@dash-frontend/providers/HoveredNodeProvider';

import {NODE_HEIGHT, NODE_WIDTH} from '../../constants/nodeSizes';
import DAGError from '../DAGError';

import DAG from './components/DAG';
import RangeSlider from './components/RangeSlider';
import styles from './DAGView.module.css';
import {MAX_SCALE_VALUE, useDAGView} from './hooks/useDAGView';

const MARKERS = [
  {id: 'end-arrow', color: '#747475'},
  {id: 'end-arrow-active', color: '#6FB3C3'},
  {id: 'end-arrow-error', color: '#E02020'},
];

type DAGViewProps = {
  dags: Dag[] | undefined;
  loading: boolean;
  error: ApolloError | undefined;
};

const DAGView: React.FC<DAGViewProps> = ({dags, loading, error}) => {
  const {
    applySliderZoom,
    minScale,
    rotateDag,
    dagDirection,
    sliderZoomValue,
    svgSize,
    zoomOut,
    skipCenterOnSelect,
    handleChangeCenterOnSelect,
  } = useDAGView(NODE_WIDTH, NODE_HEIGHT, dags, loading);

  const noDags = dags?.length === 0;

  const RotateSVG = () => {
    return (
      <FlipSVG
        className={classnames(styles.rotateSvg, {
          [styles.flipped]: dagDirection === DagDirection.RIGHT,
        })}
      />
    );
  };

  return (
    <View className={styles.view}>
      <div className={styles.topSection}>
        <div className={styles.canvasControls}>
          <RangeSlider
            min={(minScale * 100).toString()}
            max={(MAX_SCALE_VALUE * 100).toString()}
            onChange={(d: React.ChangeEvent<HTMLInputElement>) =>
              applySliderZoom(Number(d.currentTarget.value))
            }
            value={sliderZoomValue * 100}
            disabled={noDags}
          />
          <ButtonGroup>
            <Button
              className={styles.controlButton}
              buttonType="ghost"
              color="black"
              IconSVG={RotateSVG}
              disabled={noDags}
              data-testid="DAGView__flipCanvas"
              onClick={rotateDag}
            >
              Flip Canvas
            </Button>
            <Tooltip
              className={styles.tooltip}
              tooltipKey="zoomOut"
              size="large"
              placement="bottom"
              tooltipText={`Click to reset canvas, or\nuse keyboard shortcut "Shift + 2"`}
            >
              <Button
                className={styles.controlButton}
                onClick={zoomOut}
                disabled={noDags}
                data-testid="DAGView__resetCanvas"
                buttonType="ghost"
                color="black"
                IconSVG={FullscreenSVG}
              >
                Reset Canvas
              </Button>
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
              <Button
                onClick={() => {
                  handleChangeCenterOnSelect(!!skipCenterOnSelect);
                }}
                disabled={noDags}
                data-testid="DAGView__centerSelections"
                buttonType="ghost"
                color="black"
                IconSVG={!skipCenterOnSelect ? CheckboxCheckedSVG : CheckboxSVG}
              >
                Center Selections
              </Button>
            </Tooltip>
          </ButtonGroup>
        </div>
        <DAGError error={error} />
      </div>
      {noDags && (
        <EmptyState title={LETS_START_TITLE} message={NO_DAG_MESSAGE} />
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
    </View>
  );
};

export default DAGView;
