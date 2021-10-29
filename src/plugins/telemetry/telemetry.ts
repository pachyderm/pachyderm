import { JupyterFrontEnd } from '@jupyterlab/application';
import { INotebookTracker } from '@jupyterlab/notebook';
import { ITerminalTracker } from '@jupyterlab/terminal';
import { load, track } from 'rudder-sdk-js';

/**
 * TODO: This captures a lot of events. Some we might want to filter out.
 * Some events don't get captured if clicked from top level menus.
 * We'll need to figure out if menu command tracking is different,
 * or maybe even add custom tracking for them.
 */
const initCommandTracking = (app: JupyterFrontEnd): void => {
  app.commands.commandExecuted.connect((_, command) => {
    track('command', {
      id: command.id,
      // We have to copy the args to a plain object
      args: JSON.parse(JSON.stringify(command.args))
    });
  });
};

const initNotebookTracking = (notebook: INotebookTracker) => {
  // TODO: what info do we want to track from a notebook?
};

const initTerminalTracking = (terminal: ITerminalTracker) => {
  // TODO: what info do we want to track from a terminal?
};

export const init = (
  app: JupyterFrontEnd,
  notebook: INotebookTracker,
  terminal: ITerminalTracker
): void => {
  load(
    '20C6D2xFLRmyFTqtvYDEgNfwcRG',
    'https://pachyderm-dataplane.rudderstack.com'
  );
  initCommandTracking(app);
  initNotebookTracking(notebook);
  initTerminalTracking(terminal);
};
