import {JupyterLab} from '@jupyterlab/application';
import {MainMenu} from '@jupyterlab/mainmenu';
import {CommandRegistry} from '@lumino/commands';
import {showDialog} from '@jupyterlab/apputils';

import help from '../';

jest.mock('@jupyterlab/apputils');

describe('help plugin', () => {
  let app: JupyterLab;
  let mainMenu: MainMenu;
  let commands: CommandRegistry;

  beforeEach(() => {
    app = new JupyterLab();
    commands = new CommandRegistry();
    mainMenu = new MainMenu(commands);
  });

  it('should add help commands to the application', async () => {
    await help.activate(app, mainMenu);

    const appCommands = app.commands.listCommands();
    expect(appCommands.length).toEqual(2);
    expect(appCommands[0]).toEqual('jupyterlab-pachyderm:open-docs');
    expect(appCommands[1]).toEqual('jupyterlab-pachyderm:contact-support');
  });

  it('should add options to the help menu', async () => {
    await help.activate(app, mainMenu);
    const helpMenuItems = mainMenu.helpMenu.items;
    expect(helpMenuItems[1].command).toEqual('jupyterlab-pachyderm:open-docs');
    expect(helpMenuItems[2].command).toEqual(
      'jupyterlab-pachyderm:contact-support',
    );
  });

  it('should execute the open-docs command', async () => {
    window.open = jest.fn();
    await help.activate(app, mainMenu);

    app.commands.execute('jupyterlab-pachyderm:open-docs');
    expect(window.open).toHaveBeenCalledWith(
      'https://docs.pachyderm.com/latest/getting_started/',
    );
  });

  it('should execute the contact-support command', async () => {
    const mockDialog = showDialog as jest.MockedFunction<typeof showDialog>;

    await help.activate(app, mainMenu);

    app.commands.execute('jupyterlab-pachyderm:contact-support');
    expect(mockDialog).toHaveBeenCalledTimes(1);

    const {body, title} = {...mockDialog.mock.calls[0][0]};
    expect(title).toMatchSnapshot();
    expect(body).toMatchSnapshot();
  });
});
