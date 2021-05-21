import classnames from 'classnames';
import React, {useEffect, useMemo} from 'react';
import {useHistory} from 'react-router';

import {repoRoute} from '@dash-frontend/views/Project/utils/routes';

import JSONBlock from '../JSONBlock';
import {JSONBlockProps} from '../JSONBlock/JSONBlock';

import styles from './PipelineInput.module.css';

const generateInputObject = (
  jsonString: string,
  projectId: string,
  branchId: string,
) => {
  const json = JSON.parse(jsonString, (key, value) => {
    if (key === 'repo') {
      return `
        <a href=${repoRoute({
          projectId,
          repoId: value,
          branchId,
        })} id=${`__repo${value}`} class=repoLink>${value}</a>
      `.trim();
    }

    return value;
  });

  return json;
};

interface PipelineInputProps extends JSONBlockProps {
  inputString: string;
  projectId: string;
  branchId?: string;
}

const PipelineInput: React.FC<PipelineInputProps> = ({
  branchId = 'master',
  inputString,
  projectId,
  className,
  ...rest
}) => {
  const browserHistory = useHistory();

  const html = useMemo(() => {
    return JSON.stringify(
      generateInputObject(inputString, projectId, branchId),
      null,
      2,
    );
  }, [inputString, projectId, branchId]);

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
      className={classnames(styles.base, className)}
      {...rest}
    />
  );
};

export default PipelineInput;
