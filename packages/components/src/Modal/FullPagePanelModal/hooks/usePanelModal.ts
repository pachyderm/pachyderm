import {useContext} from 'react';

import ModalContext from '../contexts/ModalContext';

const usePanelModal = () => {
  const {show, onHide, onShow} = useContext(ModalContext);

  return {
    show,
    onHide,
    onShow,
  };
};

export default usePanelModal;
