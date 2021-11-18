import {JupyterLab} from '@jupyterlab/application';
import {load, track} from 'rudder-sdk-js';

import telemetry from '../';

const getNotebookAction = jest.fn(() => ({
  cell: {
    inputArea: {
      node: {
        innerText: '[1]:\npachctl version',
      },
    },
    promptNode: {
      innerText: '[1]:',
    },
  },
}));

jest.mock('rudder-sdk-js', () => ({
  load: jest.fn(),
  track: jest.fn(),
}));

jest.mock('@jupyterlab/notebook', () => {
  return {
    NotebookActions: {
      executed: {
        connect: jest.fn((cb) => cb(undefined, getNotebookAction())),
      },
    },
  };
});

describe('telemetry plugin', () => {
  let app: JupyterLab;

  beforeEach(() => {
    app = new JupyterLab();
  });

  it('should have the correct configuration', () => {
    expect(telemetry.id).toBe('jupyterlab-pachyderm:telemetry');
    expect(telemetry.autoStart).toBe(true);
  });

  it('should initialize logging', async () => {
    await telemetry.activate(app);

    expect(load).toHaveBeenCalledWith(
      '20C6D2xFLRmyFTqtvYDEgNfwcRG',
      'https://pachyderm-dataplane.rudderstack.com',
    );
  });

  it('should track app commands', async () => {
    await telemetry.activate(app);
    app.commands.addCommand('fake:command', {execute: jest.fn()});
    await app.commands.execute('fake:command', {fake: 'args'});

    expect(track).toHaveBeenCalledWith('command', {
      id: 'fake:command',
      args: {fake: 'args'},
    });
  });

  it('should track notebook commands with a prompt', async () => {
    await telemetry.activate(app);

    expect(track).toHaveBeenCalledWith('command', {
      id: 'notebook:action:executed',
      args: {
        action: 'pachctl version',
      },
    });
  });

  it('should track notebook commands without a prompt', async () => {
    getNotebookAction.mockImplementation(() => ({
      cell: {
        inputArea: {
          node: {
            innerText: 'Hello Telemetry',
          },
        },
        promptNode: {
          innerText: '',
        },
      },
    }));

    await telemetry.activate(app);

    expect(track).toHaveBeenCalledWith('command', {
      id: 'notebook:action:executed',
      args: {
        action: 'Hello Telemetry',
      },
    });
  });
});
