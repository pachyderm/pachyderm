import Modal from './components/Modal';
import ModalBody from './components/ModalBody';
import ModalError from './components/ModalError';
import ModalFooter from './components/ModalFooter';
import ModalHeader from './components/ModalHeader';

export default Object.assign(Modal, {
  Body: ModalBody,
  Footer: ModalFooter,
  Header: ModalHeader,
  Error: ModalError,
});
