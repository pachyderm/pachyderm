import {ApolloError} from '@apollo/client';
import {useState, useEffect, useCallback} from 'react';
import {useHistory} from 'react-router';

import {
  useGetClusterDefaultsQuery,
  useSetClusterDefaultsMutation,
} from '@dash-frontend/generated/hooks';
import formatJSON from '@dash-frontend/lib/formatJSON';
import {GET_CLUSTER_DEFAULTS} from '@dash-frontend/queries/GetClusterDefaults';

const useClusterConfig = () => {
  const browserHistory = useHistory();
  const [editorText, setEditorText] = useState<string>();
  const [initialEditorDoc, setInitialEditorDoc] = useState<string>();
  const [error, setError] = useState<string>();
  const [fullError, setFullError] = useState<ApolloError>();
  const [unsavedChanges, setUnsavedChanges] = useState(false);
  const [isValidJSON, setIsValidJSON] = useState(true);

  const {data: getClusterData, loading: getClusterLoading} =
    useGetClusterDefaultsQuery({
      onCompleted: (res) => {
        const resJSON = formatJSON(
          res.getClusterDefaults?.clusterDefaultsJson || '{}',
        );
        setInitialEditorDoc(resJSON);
        setEditorText(resJSON);
      },
      onError: (fullError) => {
        setFullError(fullError);
        setError('Unable to fetch cluster defaults');
      },
    });

  const checkIfChanged = useCallback(
    (editorText: string) => {
      const trimmedEditorDefaults = (editorText || '').replace(
        /[\t\s\n]+/g,
        '',
      );
      const trimmedSavedDefaults = (
        getClusterData?.getClusterDefaults?.clusterDefaultsJson || ''
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
    [
      getClusterData?.getClusterDefaults?.clusterDefaultsJson,
      getClusterLoading,
    ],
  );

  const prettifyJSON = useCallback(() => {
    const formattedString = formatJSON(editorText || '{}');
    setEditorText(formattedString);
  }, [editorText]);

  const [
    setClusterDefaultsMutation,
    {
      data: setClusterResponse,
      loading: setClusterLoading,
      error: setClusterError,
    },
  ] = useSetClusterDefaultsMutation({
    refetchQueries: [GET_CLUSTER_DEFAULTS],
    onCompleted: () => setUnsavedChanges(false),
    onError: (fullError) => {
      setFullError(fullError);
      setError('Unable to set cluster defaults');
    },
  });

  const [setClusterDryRunMutation] = useSetClusterDefaultsMutation({
    onError: (fullError) => {
      setFullError(fullError);
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
          variables: {
            args: {
              clusterDefaultsJson: editorText,
              dryRun: true,
            },
          },
          onCompleted: () => {
            setError(undefined);
            checkIfChanged(editorText);
          },
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
