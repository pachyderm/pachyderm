import {useState, useEffect, useCallback, useMemo} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import useCreatePipeline from '@dash-frontend/hooks/useCreatePipeline';
import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import {useGetClusterDefaults} from '@dash-frontend/hooks/useGetClusterDefaults';
import {useGetProjectDefaults} from '@dash-frontend/hooks/useGetProjectDefaults';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import formatJSON from '@dash-frontend/lib/formatJSON';
import {
  CREATE_PIPELINE_PATH,
  DUPLICATE_PIPELINE_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  lineageRoute,
  pipelineRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {getRandomName, useBreakpoint} from '@pachyderm/components';

const usePipelineEditor = (closeUpdateModal: () => void) => {
  const isCreating = Boolean(
    useRouteMatch({
      path: CREATE_PIPELINE_PATH,
    }),
  );
  const isDuplicating = Boolean(
    useRouteMatch({
      path: DUPLICATE_PIPELINE_PATH,
    }),
  );
  const isCreatingOrDuplicating = isCreating || isDuplicating;

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
  const [projectDefaultsJSON, setProjectDefaultsJSON] = useState<JSON>();
  const [error, setError] = useState<string>();
  const [fullError, setFullError] = useState<Error>();
  const [isValidJSON, setIsValidJSON] = useState(true);
  const {pipeline, loading: pipelineLoading} = useCurrentPipeline(!isCreating);

  // cluster defaults
  const {clusterDefaults: getClusterData, loading: clusterDefaultsLoading} =
    useGetClusterDefaults();

  const clusterDefaults = useMemo(
    () =>
      getClusterData?.clusterDefaultsJson
        ? formatJSON(getClusterData?.clusterDefaultsJson)
        : '',
    [getClusterData?.clusterDefaultsJson],
  );

  useEffect(() => {
    const clusterDefaultsJson = JSON.parse(
      getClusterData?.clusterDefaultsJson || '{}',
    );
    setClusterDefaultsJSON(clusterDefaultsJson?.createPipelineRequest);
  }, [getClusterData?.clusterDefaultsJson]);

  // project defaults
  const {projectDefaults: getProjectData, loading: projectDefaultsLoading} =
    useGetProjectDefaults({project: {name: projectId}});

  const projectDefaults = useMemo(
    () =>
      getProjectData?.projectDefaultsJson
        ? formatJSON(getProjectData?.projectDefaultsJson)
        : '',
    [getProjectData?.projectDefaultsJson],
  );

  useEffect(() => {
    const projectDefaultsJson = JSON.parse(
      getProjectData?.projectDefaultsJson || '{}',
    );
    setProjectDefaultsJSON(projectDefaultsJson?.createPipelineRequest);
  }, [getProjectData?.projectDefaultsJson]);

  const {
    createPipeline: createPipelineDryRun,
    createPipelineResponse: createDryData,
    loading: effectiveSpecLoading,
  } = useCreatePipeline({
    onSuccess: () => {
      setError(undefined);
      setFullError(undefined);
    },
    onError: (error) => {
      setError('Unable to generate an effective pipeline spec');
      setFullError(error);
    },
  });

  const {createPipeline, loading: createLoading} = useCreatePipeline({
    onSuccess: () => {
      setError(undefined);
      setFullError(undefined);
      browserHistory.push(lineageRoute({projectId}));
    },
    onError: (error) => {
      closeUpdateModal();
      setError(
        isCreatingOrDuplicating
          ? 'Unable to create pipeline'
          : 'Unable to update pipeline',
      );
      setFullError(error);
    },
  });

  useEffect(() => {
    // update effective spec from createPipelineDryRun result
    if (createDryData?.effectiveCreatePipelineRequestJson) {
      setEffectiveSpec(
        formatJSON(createDryData?.effectiveCreatePipelineRequestJson),
      );
    }
  }, [createDryData?.effectiveCreatePipelineRequestJson, effectiveSpec]);

  const addProject = useCallback(
    (editorText: string) => {
      try {
        const editorTextJSON = JSON.parse(editorText);
        if (editorTextJSON?.pipeline?.project?.name) {
          return editorText;
        } else {
          return JSON.stringify({
            ...editorTextJSON,
            pipeline: {...editorTextJSON?.pipeline, project: {name: projectId}},
          });
        }
      } catch {
        setIsValidJSON(false);
      }
    },
    [projectId],
  );

  useEffect(() => {
    // check if valid json on editorText change and trigger dry run
    try {
      const editorTextJSON = JSON.parse(editorText);
      const editorTextWithProject = addProject(editorText);
      setIsValidJSON(true);
      !pipelineLoading &&
        createPipelineDryRun({
          createPipelineRequestJson: editorTextWithProject,
          dryRun: true,
          update: !isCreatingOrDuplicating,
        });
      setEditorTextJSON(editorTextJSON);
    } catch {
      setIsValidJSON(false);
    }
  }, [
    addProject,
    createPipelineDryRun,
    editorText,
    isCreating,
    isCreatingOrDuplicating,
    pipelineLoading,
  ]);

  useEffect(() => {
    if (isMobile) {
      setSplitView(false);
    }
  }, [isMobile]);

  // Pre-load pipeline spec if updating or duplicating a pipeline
  useEffect(() => {
    if (!pipelineLoading && pipeline?.userSpecJson) {
      let newSpec = '';
      if (isDuplicating) {
        const basePipeline = JSON.parse(pipeline?.userSpecJson || '{}');
        const newPipeline = JSON.stringify({
          ...basePipeline,
          pipeline: {
            ...basePipeline.pipeline,
            name: `${basePipeline.pipeline.name}_copy`,
          },
        });
        newSpec = formatJSON(newPipeline);
      } else {
        newSpec = formatJSON(pipeline?.userSpecJson || '{}');
      }
      setEditorText(newSpec);
      setInitialDoc(newSpec);
    }
  }, [isDuplicating, pipeline?.userSpecJson, pipelineLoading]);

  const goToLineage = useCallback(
    () =>
      browserHistory.push(
        isCreatingOrDuplicating
          ? lineageRoute({projectId})
          : pipelineRoute({projectId, pipelineId}),
      ),
    [browserHistory, isCreatingOrDuplicating, pipelineId, projectId],
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
      if (effectiveSpec && !isCreatingOrDuplicating) {
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
  }, [effectiveSpec, isCreatingOrDuplicating, pipelineId]);

  return {
    editorText,
    setEditorText,
    initialDoc,
    error,
    addProject,
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
    projectDefaults,
    projectDefaultsJSON,
    projectDefaultsLoading,
    pipelineLoading,
    isCreatingOrDuplicating,
    splitView,
    setSplitView,
    isChangingProject,
    isChangingName,
    editorTextJSON,
    fullError,
  };
};

export default usePipelineEditor;
