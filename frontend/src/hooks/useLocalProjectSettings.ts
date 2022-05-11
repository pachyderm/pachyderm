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

  // listen for other instances of the hook having modified localStorage
  window.addEventListener('local-project-settings', () => {
    setValue(
      JSON.parse(
        localStorage.getItem(`pachyderm-console-${projectId}`) || '{}',
      ),
    );
  });

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
      window.dispatchEvent(new Event('local-project-settings'));
    },
    [key, projectId, settings],
  );

  return [setting, setSetting];
};

export default useLocalProjectSettings;
