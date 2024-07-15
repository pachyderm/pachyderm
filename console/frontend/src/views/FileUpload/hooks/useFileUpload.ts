import {useQueryClient} from '@tanstack/react-query';
import {
  DragEventHandler,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react';
import {useForm} from 'react-hook-form';
import {useHistory} from 'react-router';

import {useCurrentRepo} from '@dash-frontend/hooks/useCurrentRepo';
import {useLazyFetch} from '@dash-frontend/hooks/useLazyFetch';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import getAllFileEntries from '@dash-frontend/lib/getAllFileEntries';
import queryKeys from '@dash-frontend/lib/queryKeys';
import {useModal} from '@pachyderm/components';

import {GLOB_CHARACTERS, ERROR_MESSAGE} from '../lib/constants';

type FileUploadFormValues = {
  branch: string;
  path: string;
  description: string;
  files: FileList;
};

const useFileUpload = () => {
  const client = useQueryClient();
  const formCtx = useForm<FileUploadFormValues>({
    mode: 'onChange',
    defaultValues: {
      path: '/',
    },
  });

  const {
    setValue,
    formState: {errors, isValid},
    register,
    handleSubmit,
    watch,
    getValues,
    clearErrors,
  } = formCtx;

  const {repoId, projectId} = useUrlState();
  const {repo, loading: repoLoading} = useCurrentRepo();
  const {closeModal, isOpen} = useModal(true);
  const [files, setFiles] = useState<File[]>([]);
  const [maxStreamIndex, setMaxStreamIndex] = useState(10); // Limit the number of streams so that we don't overwhelm pachd
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const {goBack} = useHistory();

  const onClose = useCallback(() => {
    closeModal();
    setError('Upload cancelled');
    setTimeout(goBack, 500);
  }, [goBack, closeModal]);

  const [
    startUpload,
    {loading: startLoading, data: startData, reset: startReset},
  ] = useLazyFetch<{
    uploadId: string;
  }>({
    url: '/upload/start',
    formatResponse: async (data) => await data.json(),
    method: 'POST',
    onError: () => setError(ERROR_MESSAGE),
    onComplete: () => setError(''),
  });

  const [finishUpload, {loading: finishLoading}] = useLazyFetch<{
    commitId: string;
  }>({
    url: '/upload/finish',
    formatResponse: async (data) => await data.json(),
    method: 'POST',
    onComplete: () => {
      closeModal();
      client.invalidateQueries({
        queryKey: queryKeys.repo({
          projectId,
          repoId,
        }),
      });
      setTimeout(goBack, 500);
    },
    onError: () => setError(ERROR_MESSAGE),
  });

  const branch = watch('branch');

  const onSubmit = useCallback(
    (values: FileUploadFormValues, e?: React.BaseSyntheticEvent) => {
      e?.preventDefault();

      setLoading(true);

      startUpload({
        body: JSON.stringify({
          path: values.path,
          repo: repoId,
          branch: values.branch,
          description: values.description,
          projectId,
        }),
      });
    },
    [repoId, startUpload, projectId],
  );

  const fileDrag: DragEventHandler<HTMLDivElement> = useCallback(
    async (event) => {
      event.preventDefault();

      if (loading) return;

      try {
        const fileEntries = await getAllFileEntries(event.dataTransfer.items);

        const files = await Promise.all(
          fileEntries.map((entry) => {
            return new Promise<File>((resolve, reject) => {
              entry.file(
                (file) => {
                  resolve(file);
                },
                (err) => reject(err.message),
              );
            });
          }),
        );

        const list = new DataTransfer();
        files.forEach((file) => {
          list.items.add(file);
        });

        setValue('files', list.files, {shouldValidate: true});
        setFiles(files);
      } catch (err) {
        if (typeof err === 'string' && err.length > 0) {
          setError(err);
        } else {
          setError(ERROR_MESSAGE);
        }
      }
    },
    [setValue, loading],
  );

  const branches = useMemo(() => {
    const branches = (repo?.branches || []).map((branch) => branch.name);
    if (branches.length === 0) branches.push('master');
    return branches;
  }, [repo?.branches]);

  const handleFileCancel = useCallback(
    (index: number, success: boolean) => {
      const files = getValues('files');
      const list = new DataTransfer();

      for (let i = 0; i < files.length; i++) {
        const file = files.item(i);
        if (i === index) {
          continue;
        } else if (file) {
          list.items.add(file);
        }
      }
      if (success) {
        setMaxStreamIndex((prevValue) => prevValue - 1);
      }

      setFiles((prevValue) => {
        return prevValue.filter((_file, i) => i !== index);
      });
      setValue('files', list.files);

      if (list.files.length === 0) {
        setLoading(false);
        setError('');
        startReset();
        clearErrors('files');
      }
    },
    [setFiles, getValues, setValue, startReset, clearErrors],
  );

  const {onChange, ...fileRegister} = register('files', {
    validate: (files) => {
      if (files) {
        for (let i = 0; i < files.length; i++) {
          const file = files.item(i);
          if (file) {
            if (/[\^$\\/()|?+[\]{}><]/.test(file.name)) {
              return `Below file names cannot contain ${GLOB_CHARACTERS}. Please rename or re-upload.`;
            }
          }
        }
        return true;
      }
    },
  });

  const onChangeHandler = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setFiles(e.currentTarget.files ? Array.from(e.currentTarget.files) : []);
      onChange(e);
    },
    [onChange],
  );

  const uploadsFinished = useMemo(() => {
    return files.length > 0 && maxStreamIndex - 10 === files.length && !error;
  }, [maxStreamIndex, files, error]);

  const handleDoneClick = useCallback(() => {
    if (startData?.uploadId) {
      finishUpload({
        body: JSON.stringify({
          uploadId: startData.uploadId,
        }),
      });
    }
  }, [finishUpload, startData?.uploadId]);

  useEffect(() => {
    if (error) {
      setLoading(false);
      setMaxStreamIndex(10);
    }
  }, [error]);

  return {
    onClose,
    fileDrag,
    isOpen,
    formCtx,
    onSubmit,
    files,
    loading: repoLoading || startLoading || loading,
    branches,
    handleFileCancel,
    onChangeHandler,
    fileRegister,
    setFiles,
    errors,
    isValid,
    handleSubmit,
    setValue,
    branch,
    repo: repoId,
    uploadId: startData?.uploadId,
    error: error,
    maxStreamIndex,
    setMaxStreamIndex,
    onError: setError,
    fileNameError: errors['files'],
    uploadsFinished,
    handleDoneClick,
    finishLoading,
  };
};

export default useFileUpload;
