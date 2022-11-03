import React, {Children} from 'react';

import useFeatureFlags from '../hooks/useFeatureFlags';

import Variation from './Variation';

type ExperimentProps = {
  children: React.ReactChild[];
  name: string;
};

const Experiment: React.FC<ExperimentProps> = ({children, name}) => {
  const flags = useFeatureFlags();
  const value = flags[name];

  return (
    <>
      {Children.map(children, (child) => {
        if (
          // Pass the child through if it is not a Variation
          typeof child === 'string' ||
          typeof child === 'number' ||
          typeof child.type === 'string' ||
          child.type.name !== Variation.name
        ) {
          return child;
        }

        // The child's value has to exactly match the flag value
        if (child.props.value === value) {
          return child;
        } else {
          return null;
        }
      })}
    </>
  );
};

export default Experiment;
