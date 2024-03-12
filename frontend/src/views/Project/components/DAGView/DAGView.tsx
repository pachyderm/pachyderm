import classnames from 'classnames';
import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import {NO_DAG_MESSAGE} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import View from '@dash-frontend/components/View';
import {
  useAuthorize,
  Permission,
  ResourceType,
} from '@dash-frontend/hooks/useAuthorize';
import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DagDirection, Dags} from '@dash-frontend/lib/types';
import {
  Tooltip,
  CheckboxCheckedSVG,
  CheckboxSVG,
  FullscreenSVG,
  RotateSVG,
  Icon,
  Button,
  ButtonGroup,
  DefaultDropdown,
  OverflowSVG,
  DropdownItem,
  DownloadSVG,
  useModal,
} from '@pachyderm/components';
import useCurrentWidthAndHeight from '@pachyderm/components/hooks/useCurrentWidthAndHeight';

import {
  BOTTOM_ROUNDED_BUTTON_PATH,
  CONNECTED_EGRESS_GRADIENT_MASK,
  NODE_HEIGHT,
  NODE_WIDTH,
  PIPELINE_GRADIENT_MASK,
  REPO_GRADIENT_MASK,
  TOP_ROUNDED_BUTTON_PATH,
} from '../../constants/nodeSizes';
import CreateRepoModal from '../CreateRepoModal';
import DAGError from '../DAGError';

import DAG from './components/DAG';
import RangeSlider from './components/RangeSlider';
import styles from './DAGView.module.css';
import {useCanvasDownload} from './hooks/useCanvasDownload';
import {MAX_SCALE_VALUE, useDAGView} from './hooks/useDAGView';
import HoveredNodeProvider from './providers/HoveredNodeProvider';

const CANVAS_CONTROLS_MIN_WIDTH = 950;

const LARGE_DAG_MIN = 50;

const MARKERS = [
  {id: 'end-arrow', color: '#A4A8AE'},
  {id: 'end-arrow-active', color: '#272728'},
  {id: 'end-arrow-highlighted', color: '#007BFF'},
  {id: 'end-arrow-transferring', color: '#BDD4E0'},
];

type DAGViewProps = {
  dags: Dags | undefined;
  loading: boolean;
  error: string | undefined;
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
    graphExtents,
    projectName,
    searchParams,
  } = useDAGView(NODE_WIDTH, NODE_HEIGHT, dags, loading);
  const {isOpen: isSidebarOpen, sidebarSize} = useSidebarInfo();
  const {renderAndDownloadCanvas, downloadCanvas} = useCanvasDownload(
    searchParams,
    graphExtents,
    projectName,
  );

  const {projectId} = useUrlState();

  const {openModal, closeModal, isOpen} = useModal(false);

  const noDags = dags && dags.nodes.length === 0 && dags.links.length === 0;
  const totalNodes = dags?.nodes.length || 0;

  const {ref: headerRef, width: headerWidth} = useCurrentWidthAndHeight();
  const isResponsive = headerWidth <= CANVAS_CONTROLS_MIN_WIDTH;

  const {hasAllPermissions: hasProjectCreateRepo} = useAuthorize(
    {
      permissions: [Permission.PROJECT_CREATE_REPO],
      resource: {
        type: ResourceType.PROJECT,
        name: projectId,
      },
    },
    noDags,
  );

  const RotateSVGComponent = () => {
    return (
      <Icon
        className={classnames(styles.rotateSvg, {
          [styles.flipped]: dagDirection === DagDirection.RIGHT,
        })}
      >
        <RotateSVG />
      </Icon>
    );
  };

  const onDropdownMenuSelect = (id: string) => {
    switch (id) {
      case 'flip-canvas':
        return rotateDag();
      case 'reset-canvas':
        return zoomOut();
      case 'center-selections':
        return handleChangeCenterOnSelect(!!skipCenterOnSelect);
      case 'download-canvas':
        return downloadCanvas();
      default:
        return null;
    }
  };

  const menuItems: DropdownItem[] = [
    {
      id: 'flip-canvas',
      content: 'Flip Canvas',
      disabled: noDags,
      IconSVG: RotateSVGComponent,
    },
    {
      id: 'reset-canvas',
      content: 'Reset Canvas',
      disabled: noDags,
      IconSVG: FullscreenSVG,
    },
    {
      id: 'center-selections',
      content: 'Center Selections',
      disabled: noDags,
      IconSVG: !skipCenterOnSelect ? CheckboxCheckedSVG : CheckboxSVG,
    },
    {
      id: 'download-canvas',
      content: 'Download Canvas',
      disabled: noDags,
      IconSVG: DownloadSVG,
    },
  ];

  return (
    <View className={styles.view}>
      <div
        ref={headerRef}
        className={styles.topSection}
        style={isSidebarOpen ? {width: `calc(100% - ${sidebarSize}px)`} : {}}
      >
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
          {isResponsive && (
            <ButtonGroup>
              <div className={styles.divider} />
              <DefaultDropdown
                items={menuItems}
                onSelect={onDropdownMenuSelect}
                buttonOpts={{
                  hideChevron: true,
                  buttonType: 'ghost',
                  IconSVG: OverflowSVG,
                  color: 'black',
                }}
                menuOpts={{pin: 'left'}}
                aria-label="Open DAG controls menu"
              />
            </ButtonGroup>
          )}
          {!isResponsive && (
            <ButtonGroup>
              <Button
                className={styles.controlButton}
                buttonType="ghost"
                color="black"
                IconSVG={RotateSVGComponent}
                disabled={noDags}
                data-testid="DAGView__flipCanvas"
                onClick={rotateDag}
              >
                Flip Canvas
              </Button>
              <Tooltip
                className={styles.tooltip}
                allowedPlacements={['bottom']}
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
                allowedPlacements={['bottom']}
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
                  IconSVG={
                    !skipCenterOnSelect ? CheckboxCheckedSVG : CheckboxSVG
                  }
                >
                  Center Selections
                </Button>
              </Tooltip>
              <Button
                className={styles.controlButton}
                buttonType="ghost"
                color="black"
                IconSVG={DownloadSVG}
                disabled={noDags}
                data-testid="DAGView__downloadCanvas"
                onClick={downloadCanvas}
              >
                Download Canvas
              </Button>
            </ButtonGroup>
          )}
        </div>
        <DAGError error={error} />
      </div>
      {noDags && (
        <EmptyState
          title={''}
          message={NO_DAG_MESSAGE}
          renderButton={
            hasProjectCreateRepo ? (
              <Button onClick={openModal}>Create Your First Repo</Button>
            ) : (
              <Tooltip tooltipText="You need at least projectWriter to create a repo.">
                <Button onClick={openModal} disabled>
                  Create Your First Repo
                </Button>
              </Tooltip>
            )
          }
          linkToDocs={{
            text: 'How to create a repo on CLI',
            pathWithoutDomain: 'concepts/data-concepts/repo/',
          }}
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
                dx="2"
                dy="2"
                stdDeviation="4"
                floodColor="#d6d6d7"
              />
            </filter>
            <linearGradient
              id="green-gradient"
              x1="0%"
              y1="0%"
              x2="0%"
              y2="100%"
            >
              <stop offset="0%" stopColor="#D9E7E7" />
              <stop offset="100%" stopColor="#005D5D" />
            </linearGradient>
            <linearGradient
              id="grey-gradient"
              x1="0%"
              y1="0%"
              x2="0%"
              y2="100%"
            >
              <stop offset="0%" stopColor="#EBEBEB26" />
              <stop offset="100%" stopColor="#EBEBEB" />
            </linearGradient>
            <linearGradient
              id="purple-gradient"
              x1="0%"
              y1="0%"
              x2="0%"
              y2="100%"
            >
              <stop offset="0%" stopColor="#6929C4" />
              <stop offset="100%" stopColor="#6929c426" />
            </linearGradient>
            <path id="topRoundedButton" d={TOP_ROUNDED_BUTTON_PATH} />
            <path id="bottomRoundedButton" d={BOTTOM_ROUNDED_BUTTON_PATH} />
            <clipPath id="topRoundedButtonOnly">
              <use xlinkHref="#topRoundedButton" />
            </clipPath>
            <clipPath id="bottomRoundedButtonOnly">
              <use xlinkHref="#bottomRoundedButton" />
            </clipPath>
            <path id="repoMask" d={REPO_GRADIENT_MASK} />
            <path id="connectedEgressMask" d={CONNECTED_EGRESS_GRADIENT_MASK} />
            <path id="pipelineMask" d={PIPELINE_GRADIENT_MASK} />
          </defs>
          <g id="Dags">
            <DAG
              dagDirection={dagDirection}
              data={dags}
              rotateDag={rotateDag}
              largeDagMode={totalNodes > LARGE_DAG_MIN}
              forceFullRender={renderAndDownloadCanvas}
            />
          </g>
        </svg>
      </HoveredNodeProvider>
      {isOpen && <CreateRepoModal show={isOpen} onHide={closeModal} />}
    </View>
  );
};

export default DAGView;
