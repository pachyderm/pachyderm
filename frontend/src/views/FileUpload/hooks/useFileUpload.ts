import {useModal} from '@pachyderm/components';
import {
  DragEventHandler,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react';
import {useForm} from 'react-hook-form';
import {useHistory} from 'react-router';

import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import {useLazyFetch} from '@dash-frontend/hooks/useLazyFetch';
import useUrlState from '@dash-frontend/hooks/useUrlState';

type FileUploadFormValues = {
  branch: string;
  path: string;
  file: FileList;
};

const useFileUpload = () => {
  const formCtx = useForm<FileUploadFormValues>({
    mode: 'onChange',
  });
  const {goBack} = useHistory();

  const {repoId} = useUrlState();
  const {repo, loading: repoLoading} = useCurrentRepo();
  const {closeModal, isOpen} = useModal(true);
  const [success, setSuccess] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const {
    setValue,
    formState: {errors, isValid},
    reset,
    clearErrors,
    register,
    handleSubmit,
  } = formCtx;

  const [execFetch, {loading, error, data}] = useLazyFetch({
    url:
      process.env.NODE_ENV === 'development'
        ? 'http://localhost:3000/upload'
        : '/upload',
    formatResponse: (data) => data,
    method: 'POST',
  });

  useEffect(() => {
    if (data) setSuccess(true);
  }, [data]);

  const onSubmit = useCallback(
    (values: FileUploadFormValues, e?: React.BaseSyntheticEvent) => {
      e?.preventDefault();

      const uploadForm = new FormData();
      uploadForm.append('path', `${values.path}/${values.file[0].name}`);
      uploadForm.append('branch', values.branch);
      uploadForm.append('repo', repoId);
      uploadForm.append('file', values.file[0]);
      execFetch({
        headers: {},
        body: uploadForm,
      });
    },
    [execFetch, repoId],
  );

  const fileDrag: DragEventHandler<HTMLDivElement> = useCallback(
    async (event) => {
      event.preventDefault();
      if (loading) return;

      const entry = event.dataTransfer.items[0].webkitGetAsEntry();
      if (!entry || entry.isDirectory) return;

      const list = new DataTransfer();
      list.items.add(event.dataTransfer.files[0]);
      setSuccess(false);
      setValue('file', list.files, {shouldValidate: true});
      setFile(list.files[0]);
    },
    [setValue, loading],
  );

  const onClose = useCallback(() => {
    closeModal();

    setTimeout(goBack, 500);
  }, [goBack, closeModal]);

  const branches = useMemo(() => {
    return (repo?.branches || []).reduce(
      (acc, branch) => {
        if (branch.name !== 'master') acc.push(branch.name);
        return acc;
      },
      ['master'],
    );
  }, [repo?.branches]);

  const handleFileCancel = useCallback(() => {
    setFile(null);
    reset({file: undefined});
    clearErrors('file');
  }, [reset, clearErrors, setFile]);

  const {onChange, ...fileRegister} = register('file', {
    validate: (files) => {
      if (files && files.length === 1) {
        const file = files[0];
        if (/[\^$\\/()|?+[\]{}><]/.test(file.name)) {
          return 'File names cannot contain regex metacharacters.';
        }

        if (file.size > 1024 * 1024 * 1024) {
          return 'Please upload bigger file through terminal instead.';
        }
        return true;
      } else return false;
    },
  });

  const onChangeHandler = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setFile(e.currentTarget.files ? e.currentTarget.files[0] : null);
      setSuccess(false);
      onChange(e);
    },
    [onChange],
  );

  return {
    onClose,
    fileDrag,
    isOpen,
    formCtx,
    onSubmit,
    file,
    loading: loading || repoLoading,
    uploadError: error,
    fileError: errors['file'],
    branches,
    handleFileCancel,
    onChangeHandler,
    fileRegister,
    setFile,
    errors,
    isValid,
    handleSubmit,
    setValue,
    success,
    setSuccess,
  };
};

export default useFileUpload;
