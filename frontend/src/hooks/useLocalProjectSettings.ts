import {useState, useMemo, useCallback} from 'react';

interface LocalProjectSettingsProps {
  key: string;
  projectId: string;
}

const useLocalProjectSettings = ({
  projectId,
  key,
}: LocalProjectSettingsProps) => {
  const [settings, setValue] = useState(
    JSON.parse(localStorage.getItem(`pachyderm-console-${projectId}`) || '{}'),
  );

  const setting = useMemo(() => settings[key], [key, settings]);

  const setSetting = useCallback(
    (newValue: string) => {
      const newSettings = {
        ...settings,
        [key]: newValue,
      };
      setValue(newSettings);
      localStorage.setItem(
        `pachyderm-console-${projectId}`,
        JSON.stringify(newSettings),
      );
    },
    [key, projectId, settings],
  );

  return [setting, setSetting];
};

export default useLocalProjectSettings;
