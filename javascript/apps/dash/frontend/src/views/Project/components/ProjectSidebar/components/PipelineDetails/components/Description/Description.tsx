import React from 'react';

import styles from './Description.module.css';

interface DescriptionProps {
  term: string;
}

const Description: React.FC<DescriptionProps> = ({term, children}) => {
  return (
    <>
      <dt className={styles.term}>{term}</dt>
      <dd className={styles.description}>{children}</dd>
    </>
  );
};

export default Description;
