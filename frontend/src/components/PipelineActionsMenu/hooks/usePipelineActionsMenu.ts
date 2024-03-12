import {useMemo, useCallback} from 'react';
import {useRouteMatch, useHistory} from 'react-router';

import {PipelineState} from '@dash-frontend/api/pps';
import {
  Permission,
  ResourceType,
  useAuthorizeLazy,
} from '@dash-frontend/hooks/useAuthorize';
import {usePipelineLazy} from '@dash-frontend/hooks/usePipeline';
import {useRunCron} from '@dash-frontend/hooks/useRunCron';
import {useStartPipeline} from '@dash-frontend/hooks/useStartPipeline';
import {useStopPipeline} from '@dash-frontend/hooks/useStopPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {checkCronInputs} from '@dash-frontend/lib/checkCronInputs';
import {PROJECT_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  pipelineRoute,
  duplicatePipelineRoute,
  updatePipelineRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {
  TrashSVG,
  EditSVG,
  ActionUpdateSVG,
  DropdownItem,
  StatusPausedSVG,
  StatusCheckmarkSVG,
  CopySVG,
  DirectionsSVG,
  useNotificationBanner,
} from '@pachyderm/components';

import useDeletePipelineButton from './useDeletePipelineButton';
import useRerunPipelineButton from './useRerunPipelineButton';

const usePipelineMenu = (pipelineId: string) => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const {add: triggerNotification} = useNotificationBanner();

  const {checkPermissions, hasRepoWrite, hasRepoDelete, hasProjectCreateRepo} =
    useAuthorizeLazy({
      permissions: [
        Permission.REPO_WRITE,
        Permission.REPO_DELETE,
        Permission.PROJECT_CREATE_REPO,
      ],
      resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineId}`},
    });

  const {getPipeline, pipeline, isSpout} = usePipelineLazy({
    pipeline: {
      name: pipelineId,
      project: {name: projectId},
    },
    details: true,
  });

  const {
    openModal: openRerunPipelineModal,
    closeModal: closeRerunPipelineModal,
    isOpen: rerunPipelineModalIsOpen,
  } = useRerunPipelineButton();

  const {
    getVertices,
    modalOpen: deletePipelineModalOpen,
    setModalOpen: setDeleteModalOpen,
    canDelete,
    verticesLoading,
  } = useDeletePipelineButton(pipelineId);

  const projectMatch = !!useRouteMatch(PROJECT_PATH);

  const {stopPipeline} = useStopPipeline(() =>
    browserHistory.push(pipelineRoute({projectId, pipelineId, tabId: 'info'})),
  );

  const {startPipeline} = useStartPipeline(() =>
    browserHistory.push(pipelineRoute({projectId, pipelineId, tabId: 'info'})),
  );

  const {runCron} = useRunCron(() =>
    browserHistory.push(pipelineRoute({projectId, pipelineId, tabId: 'job'})),
  );
  const hasCronInputs = useMemo(
    () => checkCronInputs(pipeline?.details?.input),
    [pipeline],
  );

  const onDropdownMenuSelect = useCallback(
    (id: string) => {
      switch (id) {
        case 'view-dag':
          return browserHistory.push(pipelineRoute({projectId, pipelineId}));
        case 'rerun-pipeline':
          return openRerunPipelineModal();
        case 'start-pipeline':
          triggerNotification(
            `Pipeline "${pipeline?.pipeline?.name}" has been restarted`,
          );
          return startPipeline({pipeline: pipeline?.pipeline});
        case 'pause-pipeline':
          triggerNotification(
            `Pipeline "${pipeline?.pipeline?.name}" has been paused`,
          );
          return stopPipeline({pipeline: pipeline?.pipeline});
        case 'run-cron':
          triggerNotification(
            `Cron Pipeline "${pipeline?.pipeline?.name}" has been triggered`,
          );
          return runCron({pipeline: pipeline?.pipeline});
        case 'update-pipeline':
          return browserHistory.push(
            updatePipelineRoute({projectId, pipelineId}),
          );
        case 'duplicate-pipeline':
          return browserHistory.push(
            duplicatePipelineRoute({projectId, pipelineId}),
          );
        case 'delete-pipeline':
          return setDeleteModalOpen(true);
        default:
          return null;
      }
    },
    [
      browserHistory,
      openRerunPipelineModal,
      pipeline?.pipeline,
      pipelineId,
      projectId,
      runCron,
      setDeleteModalOpen,
      startPipeline,
      stopPipeline,
      triggerNotification,
    ],
  );

  const menuItems: DropdownItem[] = useMemo(() => {
    const items = [];

    const rerunPipelineTooltip = () => {
      if (isSpout) return 'Spout pipelines cannot be rerun.';
      else if (!hasRepoWrite)
        return 'You need at least repoWriter to rerun this pipeline.';

      return undefined;
    };

    const deletePipelineTooltip = () => {
      if (!hasRepoDelete) return 'You need at least repoOwner to delete this.';
      else if (!canDelete)
        return "This pipeline can't be deleted while it has downstream pipelines.";
      return undefined;
    };

    if (projectMatch) {
      items.push({
        id: 'view-dag',
        content: 'View in DAG',
        disabled: false,
        IconSVG: DirectionsSVG,
        closeOnClick: true,
      });
    }
    if (hasCronInputs) {
      items.push({
        id: 'run-cron',
        content: 'Run Cron',
        disabled: !hasRepoWrite,
        IconSVG: ActionUpdateSVG,
        closeOnClick: true,
        tooltipText: !hasRepoWrite
          ? 'You need at least repoWriter to trigger this cron pipeline'
          : undefined,
      });
    }
    items.push({
      id: 'rerun-pipeline',
      content: 'Rerun Pipeline',
      disabled: !hasRepoWrite || isSpout,
      IconSVG: ActionUpdateSVG,
      closeOnClick: true,
      tooltipText: rerunPipelineTooltip(),
    });
    if (pipeline?.state === PipelineState.PIPELINE_PAUSED) {
      items.push({
        id: 'start-pipeline',
        content: 'Restart Pipeline',
        disabled: !hasRepoWrite,
        IconSVG: StatusCheckmarkSVG,
        closeOnClick: true,
        tooltipText: !hasRepoWrite
          ? 'You need at least repoWriter to restart this pipeline.'
          : undefined,
      });
    } else {
      items.push({
        id: 'pause-pipeline',
        content: 'Pause Pipeline',
        disabled: !hasRepoWrite,
        IconSVG: StatusPausedSVG,
        closeOnClick: true,
        tooltipText: !hasRepoWrite
          ? 'You need at least repoWriter to pause this pipeline.'
          : undefined,
      });
    }
    items.push(
      {
        id: 'update-pipeline',
        content: 'Update Pipeline',
        disabled: !hasRepoWrite,
        IconSVG: EditSVG,
        closeOnClick: true,
        tooltipText: !hasRepoWrite
          ? 'You need at least repoWriter to update this pipeline.'
          : undefined,
      },
      {
        id: 'duplicate-pipeline',
        content: 'Duplicate Pipeline',
        disabled: !hasProjectCreateRepo,
        IconSVG: CopySVG,
        closeOnClick: true,
        tooltipText: !hasProjectCreateRepo
          ? 'You need at least projectWriter to create pipelines.'
          : undefined,
      },
      {
        id: 'delete-pipeline',
        content: 'Delete Pipeline',
        disabled: verticesLoading || !hasRepoDelete || !canDelete,
        IconSVG: TrashSVG,
        tooltipText: deletePipelineTooltip(),
        closeOnClick: true,
      },
    );
    return items;
  }, [
    canDelete,
    hasCronInputs,
    hasProjectCreateRepo,
    hasRepoDelete,
    hasRepoWrite,
    isSpout,
    pipeline?.state,
    projectMatch,
    verticesLoading,
  ]);

  const onMenuOpen = useCallback(() => {
    getPipeline();
    checkPermissions();
    getVertices();
  }, [checkPermissions, getPipeline, getVertices]);

  return {
    pipelineId,
    closeRerunPipelineModal,
    rerunPipelineModalIsOpen,
    deletePipelineModalOpen,
    setDeleteModalOpen,
    onDropdownMenuSelect,
    menuItems,
    onMenuOpen,
  };
};

export default usePipelineMenu;
