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

import {NODE_HEIGHT, NODE_WIDTH} from '../../constants/nodeSizes';
import CreateRepoModal from '../CreateRepoModal';
import DAGError from '../DAGError';

import DAG from './components/DAG';
import RangeSlider from './components/RangeSlider';
import styles from './DAGView.module.css';
import {useCanvasDownload} from './hooks/useCanvasDownload';
import {MAX_SCALE_VALUE, useDAGView} from './hooks/useDAGView';

const CANVAS_CONTROLS_MIN_WIDTH = 950;

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
      <DAG
        dagDirection={dagDirection}
        data={dags}
        rotateDag={rotateDag}
        svgSize={svgSize}
        forceFullRender={renderAndDownloadCanvas}
      />
      {isOpen && <CreateRepoModal show={isOpen} onHide={closeModal} />}
    </View>
  );
};

export default DAGView;
