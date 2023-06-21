import {ApolloError} from '@apollo/client';
import {Permission, ResourceType} from '@graphqlTypes';
import classnames from 'classnames';
import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import {NO_DAG_MESSAGE} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import GlobalFilter from '@dash-frontend/components/GlobalFilter';
import View from '@dash-frontend/components/View';
import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';
import {DagDirection, Dag} from '@dash-frontend/lib/types';
import HoveredNodeProvider from '@dash-frontend/providers/HoveredNodeProvider';
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
  useBreakpoint,
  DownloadSVG,
  useModal,
} from '@pachyderm/components';
import {EXTRA_LARGE} from 'constants/breakpoints';

import {NODE_HEIGHT, NODE_WIDTH} from '../../constants/nodeSizes';
import CreateRepoModal from '../CreateRepoModal';
import DAGError from '../DAGError';

import DAG from './components/DAG';
import RangeSlider from './components/RangeSlider';
import styles from './DAGView.module.css';
import {useCanvasDownload} from './hooks/useCanvasDownload';
import {MAX_SCALE_VALUE, useDAGView} from './hooks/useDAGView';

const LARGE_DAG_MIN = 50;

const MARKERS = [
  {id: 'end-arrow', color: '#747475'},
  {id: 'end-arrow-active', color: '#6FB3C3'},
  {id: 'end-arrow-error', color: '#E02020'},
];

type DAGViewProps = {
  dags: Dag[] | undefined;
  loading: boolean;
  error: ApolloError | string | undefined;
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
  const {isAuthorizedAction: createRepoIsAuthorizedAction} =
    useVerifiedAuthorization({
      permissionsList: [Permission.PROJECT_CREATE_REPO],
      resource: {type: ResourceType.PROJECT, name: projectId},
    });

  const {openModal, closeModal, isOpen} = useModal(false);

  const isResponsive = useBreakpoint(EXTRA_LARGE);
  const noDags = dags?.length === 0;
  const totalNodes = dags?.reduce((acc, val) => acc + val.nodes.length, 0) || 0;

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
        className={styles.topSection}
        style={isSidebarOpen ? {width: `calc(100% - ${sidebarSize}px`} : {}}
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
        <GlobalFilter />
      </div>
      {noDags && (
        <EmptyState
          title={''}
          message={NO_DAG_MESSAGE}
          renderButton={
            createRepoIsAuthorizedAction ? (
              <Button onClick={openModal}>Create Your First Repo</Button>
            ) : (
              <Tooltip
                tooltipKey="Create repo disabled"
                tooltipText="You need at least projectWriter to create a repo."
              >
                <span>
                  <Button
                    onClick={openModal}
                    disabled
                    className={styles.pointerEventsNone}
                  >
                    Create Your First Repo
                  </Button>
                </span>
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
                  largeDagMode={totalNodes > LARGE_DAG_MIN}
                  forceFullRender={renderAndDownloadCanvas}
                />
              );
            })}
          </g>
        </svg>
      </HoveredNodeProvider>
      {isOpen && <CreateRepoModal show={isOpen} onHide={closeModal} />}
    </View>
  );
};

export default DAGView;
