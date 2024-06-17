import isEmpty from 'lodash/isEmpty';
import {useState, useEffect, useCallback, useMemo} from 'react';
import {useHistory} from 'react-router';

import {
  Permission,
  ResourceType,
  useAuthorize,
} from '@dash-frontend/hooks/useAuthorize';
import {useGetClusterDefaults} from '@dash-frontend/hooks/useGetClusterDefaults';
import {useInspectCluster} from '@dash-frontend/hooks/useInspectCluster';
import {useSetClusterDefaults} from '@dash-frontend/hooks/useSetClusterDefaults';
import formatJSON from '@dash-frontend/lib/formatJSON';
import {useModal} from '@pachyderm/components';

const useClusterConfig = () => {
  const browserHistory = useHistory();
  const [editorText, setEditorText] = useState<string>();
  const [initialEditorDoc, setInitialEditorDoc] = useState<string>();
  const [error, setError] = useState<string>();
  const [fullError, setFullError] = useState<Error | null>();
  const [unsavedChanges, setUnsavedChanges] = useState(false);
  const [isValidJSON, setIsValidJSON] = useState(true);
  const {
    isOpen: isMetadataOpen,
    openModal: openMetadataModal,
    closeModal: closeMetadataModal,
  } = useModal();

  const {hasClusterEditMetadata} = useAuthorize({
    permissions: [Permission.CLUSTER_EDIT_CLUSTER_METADATA],
    resource: {type: ResourceType.CLUSTER, name: ''},
  });

  const {cluster} = useInspectCluster();
  const metadata = useMemo(() => cluster?.metadata, [cluster?.metadata]);
  const clusterMetadataArray = useMemo(() => {
    return !isEmpty(metadata)
      ? Object.keys(metadata).map((key) => ({
          key,
          value: metadata[key],
        }))
      : [];
  }, [metadata]);

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
    isMetadataOpen,
    openMetadataModal,
    closeMetadataModal,
    clusterMetadataArray,
    hasClusterEditMetadata,
  };
};

export default useClusterConfig;
