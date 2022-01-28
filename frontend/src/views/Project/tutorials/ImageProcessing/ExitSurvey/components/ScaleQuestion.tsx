import {RadioButton} from '@pachyderm/components';
import React from 'react';
import {FieldPath, FieldValues} from 'react-hook-form';

import styles from './ScaleQuestion.module.css';

type ScaleQuestionProps = {
  name: FieldPath<FieldValues>;
  question: string;
};

const ScaleQuestion: React.FC<ScaleQuestionProps> = ({name, question}) => {
  return (
    <>
      <div className={styles.questionHeader}>{question}</div>
      <div className={styles.radios}>
        <RadioButton
          name={name}
          value="Strongly disagree"
          className={styles.radio}
        >
          <RadioButton.Label className={styles.radioLabel}>
            Strongly disagree
          </RadioButton.Label>
        </RadioButton>
        <RadioButton name={name} value="Disagree" className={styles.radio}>
          <RadioButton.Label className={styles.radioLabel}>
            Disagree
          </RadioButton.Label>
        </RadioButton>
        <RadioButton name={name} value="Neutral" className={styles.radio}>
          <RadioButton.Label className={styles.radioLabel}>
            Neutral
          </RadioButton.Label>
        </RadioButton>
        <RadioButton name={name} value="Agree" className={styles.radio}>
          <RadioButton.Label className={styles.radioLabel}>
            Agree
          </RadioButton.Label>
        </RadioButton>
        <RadioButton
          name={name}
          value="Strongly Agree"
          className={styles.radio}
        >
          <RadioButton.Label className={styles.radioLabel}>
            Strongly agree
          </RadioButton.Label>
        </RadioButton>
      </div>
    </>
  );
};

export default ScaleQuestion;
