import {useState} from 'react';
import {useHistory} from 'react-router';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';

type useIntroductionModalProps = {
  projectId: string;
  lastPage: number;
  onClose: () => void;
};

const useIntroductionModal = ({
  projectId,
  lastPage,
  onClose,
}: useIntroductionModalProps) => {
  const routerHistory = useHistory();
  const [show, setShow] = useState(true);
  const [page, setPage] = useState(0);
  const [, setActiveTutorial] = useLocalProjectSettings({
    projectId,
    key: 'active_tutorial',
  });

  const onConfirm = () => {
    page < lastPage ? setPage(page + 1) : onStartTutorial();
  };

  const onHide = () => {
    onClose();
    setShow(false);
  };

  const onStartTutorial = () => {
    setActiveTutorial('image-processing');
    onHide();
    routerHistory.push(`/lineage/${projectId}`);
  };

  return {
    show,
    page,
    onHide,
    onConfirm,
  };
};

export default useIntroductionModal;
