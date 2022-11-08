import React from 'react';

import {
  BasicModal,
  Icon,
  EducationSVG,
  InfoSVG,
  ExternalLinkSVG,
  Link,
} from '@pachyderm/components';

import useIntroductionModal from './hooks/useIntroductionModal';
import styles from './IntroductionModal.module.css';

const pages = [
  {
    headerText: 'Welcome to Pachyderm!',
    body: (
      <div>
        <p>
          We&apos;re excited that you&apos;ve decided to try our product. Before
          you begin, please consider going through our &quot;Image
          Processing&quot; tutorial to get a hands-on feel for Pachyderm&apos;s
          capabilities!
        </p>
        <div className={styles.infoCard}>
          <div className={styles.header}>
            <h5 className={styles.headerText}>
              Image Processing at Scale with Pachyderm
            </h5>
            <p>
              Easily handle the largest structured and unstructured datasets.
            </p>
          </div>
          <div className={styles.details}>
            <div className={styles.detailsHeader}>
              <Icon className={styles.icon}>
                <EducationSVG />
              </Icon>
              <h6>
                This tutorial is an introduction to Pachyderm that explores:
              </h6>
            </div>
            <ul>
              <li>Pipelines</li>
              <li>Versioned Data</li>
              <li>Directional Acyclic Diagrams (DAGs)</li>
              <li>Data Lineage</li>
              <li>Reproducibility</li>
              <li>Incremental Scalability</li>
            </ul>
          </div>
        </div>
      </div>
    ),
    confirmText: 'Continue',
  },
  {
    headerText: 'Before you begin.',
    body: (
      <div className={styles.details}>
        <div className={styles.detailsHeader}>
          <Icon className={styles.icon}>
            <InfoSVG />
          </Icon>
          <h6>This tutorial uses real resources</h6>
        </div>
        <p>
          We will provide the files needed to run this tutorial, but you will be
          creating real repositories and pipelines. You can continue to use this
          project to play test Pachyderm after the tutorial is completed. When
          you&apos;re ready to delete everything, you can follow this guide on
          our documentation website.
        </p>
        <Link
          className={styles.link}
          small
          externalLink
          to="https://docs.pachyderm.com/latest/reference/pachctl/pachctl_delete_all/"
        >
          Pachctl delete all
          <Icon className={styles.linkSVG} small color="plum">
            <ExternalLinkSVG />
          </Icon>
        </Link>
      </div>
    ),
    confirmText: 'Start Tutorial',
  },
];

type IntroductionModalProps = {
  projectId: string;
  onClose: () => void;
};

const IntroductionModal: React.FC<IntroductionModalProps> = ({
  projectId,
  onClose,
}) => {
  const {page, onConfirm, show, onHide} = useIntroductionModal({
    projectId,
    lastPage: pages.length - 1,
    onClose,
  });

  return (
    <BasicModal
      show={show}
      loading={false}
      onHide={onHide}
      headerContent={pages[page].headerText}
      actionable
      confirmText={pages[page].confirmText}
      onConfirm={onConfirm}
      cancelText="Skip tutorial"
    >
      {pages[page].body}
    </BasicModal>
  );
};

export default IntroductionModal;
