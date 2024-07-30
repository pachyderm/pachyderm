import {useState, useEffect, useCallback} from 'react';
import {useHistory} from 'react-router';

import {useGetProjectDefaults} from '@dash-frontend/hooks/useGetProjectDefaults';
import {useSetProjectDefaults} from '@dash-frontend/hooks/useSetProjectDefaults';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import formatJSON from '@dash-frontend/lib/formatJSON';

import {lineageRoute} from '../../utils/routes';

const INITIAL_PROJECT_DEFAULTS = JSON.stringify(
  {
    createPipelineRequest: {},
  },
  null,
  2,
);

const useProjectConfig = () => {
  const browserHistory = useHistory();
  const {projectId} = useUrlState();
  const [editorText, setEditorText] = useState<string>();
  const [initialEditorDoc, setInitialEditorDoc] = useState<string>(
    INITIAL_PROJECT_DEFAULTS,
  );
  const [error, setError] = useState<string>();
  const [fullError, setFullError] = useState<Error | null>();
  const [unsavedChanges, setUnsavedChanges] = useState(false);
  const [isValidJSON, setIsValidJSON] = useState(true);

  const {
    projectDefaults,
    loading: getProjectLoading,
    error: queryError,
  } = useGetProjectDefaults({project: {name: projectId}});

  useEffect(() => {
    if (queryError) {
      setFullError(queryError);
      setError('Unable to fetch project defaults');
    }
  }, [queryError]);

  useEffect(() => {
    const emptyDefaults =
      !projectDefaults?.projectDefaultsJson ||
      projectDefaults?.projectDefaultsJson === '{}';
    const resJSON = formatJSON(
      emptyDefaults
        ? INITIAL_PROJECT_DEFAULTS
        : projectDefaults?.projectDefaultsJson || '{}',
    );
    setInitialEditorDoc(resJSON);
    setEditorText(resJSON);
  }, [projectDefaults?.projectDefaultsJson]);

  const checkIfChanged = useCallback(
    (editorText?: string) => {
      const trimmedEditorDefaults = (editorText || '').replace(
        /[\t\s\n]+/g,
        '',
      );
      const trimmedSavedDefaults = (
        projectDefaults?.projectDefaultsJson || ''
      ).replace(/[\t\s\n]+/g, '');

      if (
        !getProjectLoading &&
        editorText !== undefined &&
        trimmedEditorDefaults !== trimmedSavedDefaults
      ) {
        setUnsavedChanges(true);
      } else {
        setUnsavedChanges(false);
      }
    },
    [getProjectLoading, projectDefaults?.projectDefaultsJson],
  );

  const prettifyJSON = useCallback(() => {
    const formattedString = formatJSON(editorText || '{}');
    setEditorText(formattedString);
  }, [editorText]);

  const {
    setProjectDefaults: setProjectDefaultsMutation,
    setProjectDefaultsResponse: setProjectResponse,
    loading: setProjectLoading,
    error: setProjectError,
  } = useSetProjectDefaults({
    onSuccess: () => setUnsavedChanges(false),
    onError: (error: Error) => {
      setFullError(error);
      setError('Unable to set project defaults');
    },
  });

  const {setProjectDefaults: setProjectDryRunMutation} = useSetProjectDefaults({
    onSuccess: () => {
      setError(undefined);
      checkIfChanged(editorText);
    },
    onError: (error: Error) => {
      setFullError(error);
      setError('Unable to generate project defaults');
    },
  });

  useEffect(() => {
    // check if valid json on editorText change and trigger dry run
    if (editorText) {
      try {
        JSON.parse(editorText);
        setIsValidJSON(true);
        setProjectDryRunMutation({
          projectDefaultsJson: editorText,
          dryRun: true,
        });
      } catch {
        setIsValidJSON(false);
      }
    }
  }, [checkIfChanged, editorText, setProjectDryRunMutation]);

  const goToProject = useCallback(
    () => browserHistory.push(lineageRoute({projectId})),
    [browserHistory, projectId],
  );

  return {
    editorText,
    setEditorText,
    initialEditorDoc,
    error,
    setError,
    unsavedChanges,
    isValidJSON,
    setUnsavedChanges,
    getProjectLoading,
    setProjectDefaultsMutation,
    setProjectResponse,
    setProjectLoading,
    setProjectError,
    goToProject,
    prettifyJSON,
    fullError,
    projectId,
  };
};

export default useProjectConfig;
