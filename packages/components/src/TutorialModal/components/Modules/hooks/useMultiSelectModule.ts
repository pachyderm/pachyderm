import {useState} from 'react';

import {MultiSelectConfig} from '../../..';

type useMultiSelectModuleArgs = {
  files: Record<
    string,
    {
      name: string;
      path: string;
    }
  >;
};

export const useMultiSelectModule = ({files}: useMultiSelectModuleArgs) => {
  const [disabled, setDisabled] = useState(false);
  const [fileStates, setFileStates] = useState(() => {
    return Object.keys(files).reduce<MultiSelectConfig>((acc, key) => {
      acc[key] = {...files[key], uploaded: false, selected: false};
      return acc;
    }, {});
  });

  const onChange = (key: string) => {
    setFileStates((prevData) => {
      prevData[key].selected = !prevData[key].selected;
      return {...prevData};
    });
  };

  const setUploaded = () => {
    setFileStates((prevData) => {
      for (const file in prevData) {
        if (prevData[file].selected) prevData[file].uploaded = true;
      }
      return {...prevData};
    });
  };

  return {
    register: {
      files: fileStates,
      onChange,
      disabled,
    },
    setUploaded,
    setDisabled,
  };
};
