import { NotebookActions } from '@jupyterlab/notebook';
import { load, track } from 'rudder-sdk-js';
/**
 * TODO: This captures a lot of events. Some we might want to filter out.
 * Some events don't get captured if clicked from top level menus.
 * We'll need to figure out if menu command tracking is different,
 * or maybe even add custom tracking for them.
 */
const initCommandTracking = (app) => {
    app.commands.commandExecuted.connect((_, command) => {
        track('command', {
            id: command.id,
            // We have to copy the args to a plain object
            args: JSON.parse(JSON.stringify(command.args)),
        });
    });
};
const initNotebookTracking = () => {
    NotebookActions.executed.connect((_, action) => {
        // This transforms '[1]: pachctl version' to 'pachctl version'
        const promptText = action.cell.promptNode.innerText;
        const actionText = action.cell.inputArea.node.innerText.replace(promptText ? promptText + '\n' : '', '');
        track('command', {
            id: 'notebook:action:executed',
            args: {
                action: actionText,
            },
        });
    });
};
const initTerminalTracking = () => {
    // TODO: what info do we want to track from a terminal?
};
export const init = (app) => {
    load('20C6D2xFLRmyFTqtvYDEgNfwcRG', 'https://pachyderm-dataplane.rudderstack.com');
    initCommandTracking(app);
    initNotebookTracking();
    initTerminalTracking();
};
