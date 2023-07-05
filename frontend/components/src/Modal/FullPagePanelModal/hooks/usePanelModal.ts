import {useContext} from 'react';

import ModalContext from '../contexts/ModalContext';

const usePanelModal = () => {
  const {
    show,
    onHide,
    onShow,
    leftOpen,
    rightOpen,
    setLeftOpen,
    setRightOpen,
    hideLeftPanel,
  } = useContext(ModalContext);

  return {
    show,
    leftOpen,
    rightOpen,
    setLeftOpen,
    setRightOpen,
    onHide,
    onShow,
    hideLeftPanel,
  };
};

export default usePanelModal;
