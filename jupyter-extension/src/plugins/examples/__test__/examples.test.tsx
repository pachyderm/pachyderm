import {JupyterLab} from '@jupyterlab/application';
import {LauncherModel} from '@jupyterlab/launcher';
import examples from '../';

describe('examples plugin', () => {
  let app: JupyterLab;
  let launcher: LauncherModel;

  beforeEach(() => {
    app = new JupyterLab();
    launcher = new LauncherModel();
  });

  it('should add example if the file is found', async () => {
    fetchMock.mockResponse('');
    await examples.activate(app, launcher);

    const appCommands = app.commands.listCommands();
    expect(appCommands).toHaveLength(2);
    expect(appCommands[0]).toBe('jupyterlab-pachyderm:open-example-intro');
    expect(appCommands[1]).toBe('jupyterlab-pachyderm:open-example-mount');

    const launcherItems = launcher.items();
    expect(launcherItems.next()).toEqual({
      category: 'Pachyderm Examples',
      command: 'jupyterlab-pachyderm:open-example-intro',
      rank: 1,
    });
    expect(launcherItems.next()).toEqual({
      category: 'Pachyderm Examples',
      command: 'jupyterlab-pachyderm:open-example-mount',
      rank: 2,
    });
    expect(launcherItems.next()).toBeUndefined();
  });

  it('should not add example if the file is not found', async () => {
    fetchMock.mockReject();
    await examples.activate(app, launcher);
    expect(app.commands.listCommands()).toHaveLength(0);
    expect(launcher.items().next()).toBeUndefined();
  });
});
