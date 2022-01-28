import {useModal} from '@pachyderm/components';
import {useEffect, useCallback} from 'react';
import {useForm, SubmitHandler} from 'react-hook-form';

import useAccount from '@dash-frontend/hooks/useAccount';
import {useLazyFetch} from '@dash-frontend/hooks/useLazyFetch';

const FORM_URL =
  'https://api.hsforms.com/submissions/v3/integration/submit/4751021/82d19a69-40d7-454d-a81a-4c3d2dd25e0e';

type AnswerScale =
  | 'Strongly disagree'
  | 'Disagree'
  | 'Neutral'
  | 'Agree'
  | 'Strongly agree';

type FormValues = {
  // eslint-disable-next-line @typescript-eslint/naming-convention
  exit_survey___q1: AnswerScale;
  // eslint-disable-next-line @typescript-eslint/naming-convention
  exit_survey___q2: AnswerScale;
};

const useExitSurvey = (onTutorialClose: () => void) => {
  const {isOpen, closeModal} = useModal(true);
  const {account} = useAccount();
  const formContext = useForm<FormValues>();
  const {watch} = formContext;
  const watchFields = watch();

  const [submitForm, {loading: updating, data, error}] = useLazyFetch<void>({
    method: 'POST',
    url: FORM_URL,
    formatResponse: async (res) => res.json(),
    credentials: 'same-origin',
  });

  const onSubmit: SubmitHandler<FormValues> = (fields) => {
    submitForm({
      body: JSON.stringify({
        context: {
          pageName: 'Image Processing Tutorial Exit Survey',
          pageUri: window.location.origin,
        },
        submittedAt: Date.now().toString(),
        fields: [
          ...Object.entries(fields).map(([name, value]) => ({
            name,
            value,
          })),
          {
            name: 'email',
            value: account?.email,
          },
        ],
      }),
      method: 'POST',
    });
  };

  const onClose = useCallback(() => {
    onTutorialClose();
    closeModal();
  }, [onTutorialClose, closeModal]);

  useEffect(() => {
    if (data) {
      onClose();
    }
  }, [data, onClose]);

  return {
    isOpen,
    onSubmit,
    formContext,
    updating,
    disabled:
      !Object.values(watchFields).length ||
      Object.values(watchFields).some((f) => !f),
    onClose,
    error,
  };
};

export default useExitSurvey;
