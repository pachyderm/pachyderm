import {useCallback} from 'react';

import {useDeletePipeline} from '@dash-frontend/hooks/useDeletePipeline';
import {useDeleteRepo} from '@dash-frontend/hooks/useDeleteRepo';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import usePipelines from '@dash-frontend/hooks/usePipelines';
import useRepos from '@dash-frontend/hooks/useRepos';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import useOnCloseTutorial from '../../hooks/useCloseTutorial';

const useImageProcessingCleanup = () => {
  const {projectId} = useUrlState();
  const {repos} = useRepos({projectId});
  const {pipelines} = usePipelines({projectIds: [projectId]});
  const onCloseTutorial = useOnCloseTutorial({clearProgress: true});
  const [tutorialProgress] = useLocalProjectSettings({
    projectId,
    key: 'tutorial_progress',
  });
  const tutorialId =
    tutorialProgress && tutorialProgress['image-processing']?.tutorialId;

  const {
    deletePipeline: deletePipelineMontage,
    loading: deletePipelineMontageLoading,
    error: deletePipelineMontageError,
  } = useDeletePipeline(`montage_${tutorialId}`);
  const {
    deletePipeline: deletePipelineEdges,
    loading: deletePipelineEdgesLoading,
    error: deletePipelineEdgesError,
  } = useDeletePipeline(`edges_${tutorialId}`);
  const {
    deleteRepo: deleteRepoImages,
    loading: deleteRepoImagesLoading,
    error: deleteRepoImagesError,
  } = useDeleteRepo(`images_${tutorialId}`);

  const cleanupImageProcessing = useCallback(async () => {
    const callDeletePipelineMontage = () =>
      deletePipelineMontage({
        variables: {
          args: {
            name: `montage_${tutorialId}`,
            projectId,
          },
        },
      });

    const callDeletePipelineEdges = () =>
      deletePipelineEdges({
        variables: {
          args: {
            name: `edges_${tutorialId}`,
            projectId,
          },
        },
      });

    const callDeleteRepoImages = () =>
      deleteRepoImages({
        variables: {
          args: {
            repo: {name: `images_${tutorialId}`},
            projectId,
          },
        },
      });

    if (
      pipelines?.find((pipeline) => `montage_${tutorialId}` === pipeline?.name)
    ) {
      callDeletePipelineMontage()
        .then(callDeletePipelineEdges)
        .then(callDeleteRepoImages);
    } else if (
      pipelines?.find((pipeline) => `edges_${tutorialId}` === pipeline?.name)
    ) {
      callDeletePipelineEdges().then(callDeleteRepoImages);
    } else if (repos?.find((repo) => `images_${tutorialId}` === repo?.name)) {
      callDeleteRepoImages();
    }
    onCloseTutorial('image-processing');
  }, [
    deletePipelineEdges,
    deletePipelineMontage,
    deleteRepoImages,
    onCloseTutorial,
    pipelines,
    projectId,
    repos,
    tutorialId,
  ]);

  return {
    cleanupImageProcessing,
    loading:
      deletePipelineMontageLoading ||
      deletePipelineEdgesLoading ||
      deleteRepoImagesLoading,
    error:
      deletePipelineMontageError ||
      deletePipelineEdgesError ||
      deleteRepoImagesError,
  };
};

export default useImageProcessingCleanup;
