import {JupyterFrontEnd} from '@jupyterlab/application';
import debounce from 'lodash/debounce';
import {load, track} from 'rudder-sdk-js';

export const CLICK_TIMEOUT = 500;

const handleClick = debounce(
  (evt: Event) => {
    const element = evt.target as HTMLElement;
    const clickId = element.getAttribute('data-testid');

    if (clickId) {
      track('notebook:click', {clickId});
    }
  },
  CLICK_TIMEOUT,
  {
    leading: true,
    trailing: false,
  },
);

const initClickTracking = () => {
  if (window.document.onclick === handleClick) {
    return;
  }

  window.document.addEventListener('click', handleClick);
};

export const init = (app: JupyterFrontEnd): void => {
  load(
    '20C6D2xFLRmyFTqtvYDEgNfwcRG',
    'https://pachyderm-dataplane.rudderstack.com',
  );
  initClickTracking();
};
