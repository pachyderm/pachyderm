import classnames from 'classnames';
import noop from 'lodash/noop';
import React, {useState} from 'react';
import BootstrapModalFooter from 'react-bootstrap/ModalFooter';

import {Button} from './../../Button';
import {ButtonLink} from './../../ButtonLink';
import {Group} from './../../Group';
import {Modal} from './../../Modal';
import styles from './WizardModal.module.css';

type WizardModalProps = {
  show: boolean;
  onHide?: () => void;
  onShow?: () => void;
  headerContent: React.ReactNode[];
  loading?: boolean;
  className?: string;
  errorMessage?: string;
  modalContent: React.ReactNode[];
};

const WizardModal: React.FC<WizardModalProps> = ({
  show,
  onHide = noop,
  onShow = noop,
  headerContent,
  loading = true,
  className,
  errorMessage = '',
  modalContent,
}) => {
  const [modalIndex, setModalIndex] = useState(0);

  return (
    <Modal
      show={show}
      onHide={onHide}
      onShow={onShow}
      className={className}
      pinTop={true}
    >
      {errorMessage && <Modal.Error>{errorMessage}</Modal.Error>}

      <Modal.Header
        onHide={onHide}
        error={Boolean(errorMessage)}
        loading={loading}
      >
        {headerContent[modalIndex]}
      </Modal.Header>

      <Modal.Body>{modalContent[modalIndex]}</Modal.Body>

      <BootstrapModalFooter className={styles.footer}>
        <Group spacing={32}>
          {modalIndex > 0 && (
            <ButtonLink
              onClick={() => setModalIndex(modalIndex - 1)}
              type="button"
            >
              Previous
            </ButtonLink>
          )}
          {modalIndex === modalContent.length - 1 ? (
            <Button onClick={onHide}>Done</Button>
          ) : (
            <Button onClick={() => setModalIndex(modalIndex + 1)}>
              Ok, Next
            </Button>
          )}
        </Group>
        <div className={styles.dots}>
          {modalContent.map((_content, index) => (
            <span
              key={index}
              className={classnames(styles.dot, {
                [styles.active]: index === modalIndex,
              })}
            >
              ⬤
            </span>
          ))}
        </div>
      </BootstrapModalFooter>
    </Modal>
  );
};

export default WizardModal;
