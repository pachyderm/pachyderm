import {GRPC_MAX_MESSAGE_LENGTH} from '@dash-backend/lib/constants';
import {useCallback, useEffect, useReducer, useState} from 'react';

import {useFileDetails} from '@dash-frontend/components/CodePreview';
import {ERROR_MESSAGE} from '@dash-frontend/views/FileUpload/lib/constants';

import useCancelFile from './useCancelFile';

type UseFileCardProps = {
  file: File;
  uploadId?: string;
  maxStreamIndex: number;
  index: number;
  onComplete: React.Dispatch<React.SetStateAction<number>>;
  handleFileCancel: (index: number, success: boolean) => void;
  onError: React.Dispatch<React.SetStateAction<string>>;
  uploadError: boolean;
};

const CHUNK_SIZE = GRPC_MAX_MESSAGE_LENGTH;

type UploadState = {
  loading: boolean;
  error: boolean;
  progress: number;
  success: boolean;
};

type UploadAction =
  | {type: 'STARTED'}
  | {type: 'PROGRESS'; payload: number}
  | {type: 'ERROR'}
  | {type: 'RESET'}
  | {type: 'SUCCESS'};

const useFileCard = ({
  file,
  uploadId = '',
  maxStreamIndex,
  onComplete,
  index,
  handleFileCancel,
  onError,
  uploadError,
}: UseFileCardProps) => {
  const [controller, setController] = useState(new AbortController());
  const {fileDetails} = useFileDetails(file.name);

  const initialState: UploadState = {
    loading: false,
    error: false,
    progress: 0,
    success: false,
  };

  const [state, dispatch] = useReducer(
    (state: UploadState, action: UploadAction) => {
      switch (action.type) {
        case 'STARTED':
          return {...initialState, loading: true};
        case 'PROGRESS':
          return {
            ...initialState,
            loading: true,
            progress: action.payload,
          };
        case 'ERROR':
          return {...initialState, error: true};
        case 'RESET':
          return initialState;
        case 'SUCCESS':
          return {...initialState, success: true};
        default:
          return state;
      }
    },
    initialState,
  );

  const fileMajorType = fileDetails.icon;

  const onCancel = useCallback(() => {
    if (!uploadError) handleFileCancel(index, state.success);
  }, [uploadError, handleFileCancel, index, state.success]);

  const {cancelFile, loading: cancelLoading} = useCancelFile({
    onCancel,
    onError: () => {
      dispatch({type: 'ERROR'});
      onError('Cancellation failed');
    },
    fileId: file.name,
    uploadId,
  });

  const uploadFile = useCallback(() => {
    let start = 0;
    const controller = new AbortController();
    setController(controller);
    const uploadChunk = async (chunkForm: FormData) => {
      try {
        await fetch('/upload', {
          credentials: 'include',
          mode: 'cors',
          headers: {},
          method: 'POST',
          signal: controller.signal,
          body: chunkForm,
        });
        start += CHUNK_SIZE;
        if (start < file.size) {
          createChunk(start);
          dispatch({
            type: 'PROGRESS',
            payload: Math.round((start / file.size) * 100),
          });
        } else {
          dispatch({type: 'SUCCESS'});
          onComplete((preValue) => preValue + 1);
        }
      } catch (err) {
        if (err instanceof Error) {
          if (err.name === 'AbortError') {
            await cancelFile();
          } else {
            dispatch({type: 'ERROR'});
            onError(ERROR_MESSAGE);
          }
        } else {
          dispatch({type: 'ERROR'});
          onError(ERROR_MESSAGE);
        }
      }
    };

    const createChunk = (start: number) => {
      const chunkEnd = Math.min(start + CHUNK_SIZE, file.size);
      const chunk = file.slice(start, chunkEnd);
      const chunkForm = new FormData();
      chunkForm.append('uploadId', uploadId);
      chunkForm.append('fileName', file.name);
      chunkForm.append(
        'chunkTotal',
        Math.ceil(file.size / CHUNK_SIZE).toString(),
      );
      chunkForm.append(
        'currentChunk',
        Math.ceil(chunkEnd / CHUNK_SIZE).toString(),
      );
      chunkForm.append('file', chunk);
      uploadChunk(chunkForm);
    };
    dispatch({type: 'STARTED'});
    createChunk(start);
  }, [uploadId, file, onComplete, cancelFile, onError]);

  const cancel = useCallback(async () => {
    if (!uploadId || index > maxStreamIndex) {
      handleFileCancel(index, state.success);
    } else if (!state.loading) {
      await cancelFile();
    } else {
      controller.abort();
    }
  }, [
    controller,
    state.loading,
    index,
    maxStreamIndex,
    cancelFile,
    handleFileCancel,
    uploadId,
    state.success,
  ]);

  // determine when to upload file
  useEffect(() => {
    if (
      uploadId &&
      index <= maxStreamIndex &&
      !state.success &&
      !state.loading &&
      !uploadError
    ) {
      uploadFile();
    }
  }, [
    uploadId,
    state.success,
    state.loading,
    index,
    maxStreamIndex,
    uploadFile,
    uploadError,
  ]);

  // reset state and abort upload on general upload errors
  useEffect(() => {
    if (uploadError) {
      dispatch({type: 'ERROR'});
      controller.abort();
    }
  }, [uploadError, controller]);

  useEffect(() => {
    if (/[\^$\\/()|?+[\]{}><]/.test(file.name)) {
      dispatch({type: 'ERROR'});
    } else {
      dispatch({type: 'RESET'});
    }
  }, [file.name]);

  return {
    fileMajorType,
    cancel,
    cancelLoading,
    ...state,
  };
};

export default useFileCard;
