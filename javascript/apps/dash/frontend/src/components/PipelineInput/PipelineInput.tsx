import React, {useEffect, useMemo} from 'react';
import {useHistory} from 'react-router';

import {repoRoute} from '@dash-frontend/views/Project/utils/routes';

import JSONBlock from '../JSONBlock';
import {JSONBlockProps} from '../JSONBlock/JSONBlock';

const generateInputObject = (jsonString: string, projectId: string) => {
  const json = JSON.parse(jsonString, (key, value) => {
    if (key === 'repo') {
      return `
        <a href=${repoRoute({
          projectId,
          repoId: value,
        })} id=${`__repo${value}`}>${value}</a>
      `.trim();
    }

    return value;
  });

  return json;
};

interface PipelineInputProps extends JSONBlockProps {
  inputString: string;
  projectId: string;
}

const PipelineInput: React.FC<PipelineInputProps> = ({
  inputString,
  projectId,
  ...rest
}) => {
  const browserHistory = useHistory();

  const html = useMemo(() => {
    return JSON.stringify(generateInputObject(inputString, projectId), null, 2);
  }, [inputString, projectId]);

  useEffect(() => {
    const links = Array.from(
      document.querySelectorAll<HTMLAnchorElement>(`[id^='__repo'`),
    );

    const handleClick = (e: MouseEvent) => {
      e.preventDefault();

      if (e.target instanceof HTMLAnchorElement) {
        browserHistory.push(e.target.pathname);
      }
    };

    links.forEach((link) => {
      link.addEventListener('click', handleClick);
    });

    return () => {
      links.forEach((link) => {
        link.removeEventListener('click', handleClick);
      });
    };
    // NOTE: eventListeners need to be re-attached when
    // inputString and projectId change
  }, [browserHistory, inputString, projectId]);

  return (
    <JSONBlock
      dangerouslySetInnerHTML={{
        // eslint-disable-next-line @typescript-eslint/naming-convention
        __html: html,
      }}
      {...rest}
    />
  );
};

export default PipelineInput;
