import isEqual from 'lodash/isEqual';
import {useState, useCallback, useEffect} from 'react';

interface LocalProjectSettingsProps {
  key: string;
  projectId: string;
}

const useLocalProjectSettings = ({
  projectId,
  key,
}: LocalProjectSettingsProps) => {
  const [setting, setValue] = useState(
    JSON.parse(localStorage.getItem(`pachyderm-console-${projectId}`) || '{}')[
      key
    ],
  );

  // listen for other instances of the hook having modified localStorage
  useEffect(() => {
    const updateSettings = () => {
      const newSettings = JSON.parse(
        localStorage.getItem(`pachyderm-console-${projectId}`) || '{}',
      );

      if (!isEqual(setting, newSettings[key])) {
        setValue(newSettings[key]);
      }
    };
    window.addEventListener(
      `local-project-settings-${projectId}`,
      updateSettings,
    );
    return () =>
      window.removeEventListener(
        `local-project-settings-${projectId}`,
        updateSettings,
      );
  }, [key, projectId, setting]);

  const setSetting = useCallback(
    (newValue: string) => {
      const oldSettings = JSON.parse(
        localStorage.getItem(`pachyderm-console-${projectId}`) || '{}',
      );
      const newSettings = {
        ...oldSettings,
        [key]: newValue,
      };
      setValue(newValue);
      localStorage.setItem(
        `pachyderm-console-${projectId}`,
        JSON.stringify(newSettings),
      );
      window.dispatchEvent(new Event(`local-project-settings-${projectId}`));
    },
    [key, projectId],
  );

  return [setting, setSetting];
};

export default useLocalProjectSettings;
