import {useState, useEffect, useCallback} from 'react';
import {useHistory} from 'react-router';

import {useGetClusterDefaults} from '@dash-frontend/hooks/useGetClusterDefaults';
import {useSetClusterDefaults} from '@dash-frontend/hooks/useSetClusterDefaults';
import formatJSON from '@dash-frontend/lib/formatJSON';

const useClusterConfig = () => {
  const browserHistory = useHistory();
  const [editorText, setEditorText] = useState<string>();
  const [initialEditorDoc, setInitialEditorDoc] = useState<string>();
  const [error, setError] = useState<string>();
  const [fullError, setFullError] = useState<Error | null>();
  const [unsavedChanges, setUnsavedChanges] = useState(false);
  const [isValidJSON, setIsValidJSON] = useState(true);

  const {
    clusterDefaults,
    loading: getClusterLoading,
    error: queryError,
  } = useGetClusterDefaults();

  useEffect(() => {
    if (queryError) {
      setFullError(queryError);
      setError('Unable to fetch cluster defaults');
    }
  }, [queryError]);

  useEffect(() => {
    const resJSON = formatJSON(clusterDefaults?.clusterDefaultsJson || '{}');
    setInitialEditorDoc(resJSON);
    setEditorText(resJSON);
  }, [clusterDefaults]);

  const checkIfChanged = useCallback(
    (editorText?: string) => {
      const trimmedEditorDefaults = (editorText || '').replace(
        /[\t\s\n]+/g,
        '',
      );
      const trimmedSavedDefaults = (
        clusterDefaults?.clusterDefaultsJson || ''
      ).replace(/[\t\s\n]+/g, '');

      if (
        !getClusterLoading &&
        editorText !== undefined &&
        trimmedEditorDefaults !== trimmedSavedDefaults
      ) {
        setUnsavedChanges(true);
      } else {
        setUnsavedChanges(false);
      }
    },
    [clusterDefaults?.clusterDefaultsJson, getClusterLoading],
  );

  const prettifyJSON = useCallback(() => {
    const formattedString = formatJSON(editorText || '{}');
    setEditorText(formattedString);
  }, [editorText]);

  const {
    setClusterDefaults: setClusterDefaultsMutation,
    setClusterDefaultsResponse: setClusterResponse,
    loading: setClusterLoading,
    error: setClusterError,
  } = useSetClusterDefaults({
    onSuccess: () => setUnsavedChanges(false),
    onError: (error: Error) => {
      setFullError(error);
      setError('Unable to set cluster defaults');
    },
  });

  const {setClusterDefaults: setClusterDryRunMutation} = useSetClusterDefaults({
    onSuccess: () => {
      setError(undefined);
      checkIfChanged(editorText);
    },
    onError: (error: Error) => {
      setFullError(error);
      setError('Unable to generate cluster defaults');
    },
  });

  useEffect(() => {
    // check if valid json on editorText change and trigger dry run
    if (editorText) {
      try {
        JSON.parse(editorText);
        setIsValidJSON(true);
        setClusterDryRunMutation({
          clusterDefaultsJson: editorText,
          dryRun: true,
        });
      } catch {
        setIsValidJSON(false);
      }
    }
  }, [checkIfChanged, editorText, setClusterDryRunMutation]);

  const goToLanding = useCallback(
    () => browserHistory.push('/'),
    [browserHistory],
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
    getClusterLoading,
    setClusterDefaultsMutation,
    setClusterResponse,
    setClusterLoading,
    setClusterError,
    goToLanding,
    prettifyJSON,
    fullError,
  };
};

export default useClusterConfig;
