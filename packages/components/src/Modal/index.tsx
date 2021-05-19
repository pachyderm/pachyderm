import BasicModal from './BasicModal';
import Modal from './components/Modal';
import ModalBody from './components/ModalBody';
import ModalError from './components/ModalError';
import ModalFooter from './components/ModalFooter';
import ModalHeader from './components/ModalHeader';
import FormModal from './FormModal';
import FullPageModal from './FullPageModal';
import useModal from './hooks/useModal';

export {useModal};
export {BasicModal};
export {FullPageModal};
export {FormModal};

export default Object.assign(Modal, {
  Body: ModalBody,
  Footer: ModalFooter,
  Header: ModalHeader,
  Error: ModalError,
});
