import {ApolloError} from '@apollo/client';
import {useState, useEffect, useCallback, useMemo} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import {CLUSTER_DEFAULTS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {
  useCreatePipelineV2Mutation,
  useGetClusterDefaultsQuery,
} from '@dash-frontend/generated/hooks';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import formatJSON from '@dash-frontend/lib/formatJSON';
import {CREATE_PIPELINE_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  lineageRoute,
  pipelineRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {getRandomName, useBreakpoint} from '@pachyderm/components';

const usePipelineEditor = (closeUpdateModal: () => void) => {
  const {projectId, pipelineId} = useUrlState();
  const browserHistory = useHistory();
  const initialEditorDoc = JSON.stringify(
    {
      pipeline: {
        name: getRandomName(),
        project: {
          name: projectId,
        },
      },
      transform: {},
    },
    null,
    2,
  );

  const [splitView, setSplitView] = useState(true);
  const isMobile = useBreakpoint(1000);
  const [initialDoc, setInitialDoc] = useState(initialEditorDoc);
  const [editorText, setEditorText] = useState(initialEditorDoc);
  const [effectiveSpec, setEffectiveSpec] = useState<string>();
  const [editorTextJSON, setEditorTextJSON] = useState<JSON>();
  const [clusterDefaultsJSON, setClusterDefaultsJSON] = useState<JSON>();
  const [error, setError] = useState<string>();
  const [fullError, setFullError] = useState<ApolloError>();
  const [isValidJSON, setIsValidJSON] = useState(true);
  const isCreating = Boolean(
    useRouteMatch({
      path: CREATE_PIPELINE_PATH,
    }),
  );
  const {pipeline, loading: pipelineLoading} = useCurrentPipeline({
    skip: isCreating,
    fetchPolicy: 'network-only',
  });

  const {data: getClusterData, loading: clusterDefaultsLoading} =
    useGetClusterDefaultsQuery({
      onCompleted: () => {
        setError(undefined);
        setFullError(undefined);
      },
      pollInterval: CLUSTER_DEFAULTS_POLL_INTERVAL_MS,
    });

  const clusterDefaults = useMemo(
    () =>
      formatJSON(
        getClusterData?.getClusterDefaults?.clusterDefaultsJson || '{}',
      ),
    [getClusterData?.getClusterDefaults?.clusterDefaultsJson],
  );

  useEffect(() => {
    const clusterDefaultsJson = JSON.parse(
      getClusterData?.getClusterDefaults?.clusterDefaultsJson || '{}',
    );
    setClusterDefaultsJSON(clusterDefaultsJson?.createPipelineRequest);
  }, [getClusterData?.getClusterDefaults?.clusterDefaultsJson]);

  const [
    createPipelineDryRun,
    {data: createDryData, loading: effectiveSpecLoading},
  ] = useCreatePipelineV2Mutation({
    onCompleted: () => {
      setError(undefined);
      setFullError(undefined);
    },
    onError: (error) => {
      setError('Unable to generate an effective pipeline spec');
      setFullError(error);
    },
  });

  const [createPipeline, {loading: createLoading}] =
    useCreatePipelineV2Mutation({
      onCompleted: () => {
        setError(undefined);
        setFullError(undefined);
        browserHistory.push(lineageRoute({projectId}));
      },
      onError: (error) => {
        closeUpdateModal();
        setError(
          isCreating
            ? 'Unable to create pipeline'
            : 'Unable to update pipeline',
        );
        setFullError(error);
      },
    });

  useEffect(() => {
    // update effective spec from createPipelineDryRun result
    if (createDryData?.createPipelineV2.effectiveCreatePipelineRequestJson) {
      setEffectiveSpec(
        formatJSON(
          createDryData?.createPipelineV2.effectiveCreatePipelineRequestJson,
        ),
      );
    }
  }, [
    createDryData?.createPipelineV2.effectiveCreatePipelineRequestJson,
    effectiveSpec,
  ]);

  useEffect(() => {
    // check if valid json on editorText change and trigger dry run
    try {
      const editorTextJSON = JSON.parse(editorText);
      setIsValidJSON(true);
      !pipelineLoading &&
        createPipelineDryRun({
          variables: {
            args: {
              createPipelineRequestJson: editorText,
              dryRun: true,
              update: !isCreating,
            },
          },
        });
      setEditorTextJSON(editorTextJSON);
    } catch {
      setIsValidJSON(false);
    }
  }, [createPipelineDryRun, editorText, isCreating, pipelineLoading]);

  useEffect(() => {
    if (isMobile) {
      setSplitView(false);
    }
  }, [isMobile]);

  // Pre-load pipeline spec if updating a pipeline
  useEffect(() => {
    if (!pipelineLoading && pipeline?.userSpecJson) {
      const newSpec = formatJSON(pipeline?.userSpecJson || '{}');
      setEditorText(newSpec);
      setInitialDoc(newSpec);
    }
  }, [pipeline?.userSpecJson, pipelineLoading]);

  const goToLineage = useCallback(
    () =>
      browserHistory.push(
        isCreating
          ? lineageRoute({projectId})
          : pipelineRoute({projectId, pipelineId}),
      ),
    [browserHistory, isCreating, pipelineId, projectId],
  );

  const prettifyJSON = useCallback(() => {
    const formattedString = formatJSON(editorText || '{}');
    setEditorText(formattedString);
  }, [editorText]);

  const isChangingProject = useMemo(() => {
    try {
      if (effectiveSpec) {
        const spec = JSON.parse(effectiveSpec);
        const targetProject = spec.pipeline.project.name;

        if (targetProject && targetProject !== projectId) {
          return targetProject;
        }
      }
      return undefined;
    } catch {
      return undefined;
    }
  }, [effectiveSpec, projectId]);

  const isChangingName = useMemo(() => {
    try {
      if (effectiveSpec && !isCreating) {
        const spec = JSON.parse(effectiveSpec);
        const pipelineName = spec.pipeline.name;

        if (pipelineName && pipelineName !== pipelineId) {
          return pipelineName;
        }
      }
      return undefined;
    } catch {
      return undefined;
    }
  }, [effectiveSpec, isCreating, pipelineId]);

  return {
    editorText,
    setEditorText,
    initialDoc,
    error,
    createPipeline,
    createLoading,
    goToLineage,
    prettifyJSON,
    isValidJSON,
    effectiveSpec,
    effectiveSpecLoading,
    clusterDefaults,
    clusterDefaultsJSON,
    clusterDefaultsLoading,
    pipelineLoading,
    isCreating,
    splitView,
    setSplitView,
    isChangingProject,
    isChangingName,
    editorTextJSON,
    fullError,
  };
};

export default usePipelineEditor;
