var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { JupyterLab } from '@jupyterlab/application';
import { MainMenu } from '@jupyterlab/mainmenu';
import { CommandRegistry } from '@lumino/commands';
import { showDialog } from '@jupyterlab/apputils';
import help from '../';
jest.mock('@jupyterlab/apputils');
describe('help plugin', () => {
    let app;
    let mainMenu;
    let commands;
    beforeEach(() => {
        app = new JupyterLab();
        commands = new CommandRegistry();
        mainMenu = new MainMenu(commands);
    });
    it('should add help commands to the application', () => __awaiter(void 0, void 0, void 0, function* () {
        yield help.activate(app, mainMenu);
        const appCommands = app.commands.listCommands();
        expect(appCommands.length).toEqual(2);
        expect(appCommands[0]).toEqual('jupyterlab-pachyderm:open-docs');
        expect(appCommands[1]).toEqual('jupyterlab-pachyderm:contact-support');
    }));
    it('should add options to the help menu', () => __awaiter(void 0, void 0, void 0, function* () {
        yield help.activate(app, mainMenu);
        const helpMenuItems = mainMenu.helpMenu.items;
        expect(helpMenuItems[1].command).toEqual('jupyterlab-pachyderm:open-docs');
        expect(helpMenuItems[2].command).toEqual('jupyterlab-pachyderm:contact-support');
    }));
    it('should execute the open-docs command', () => __awaiter(void 0, void 0, void 0, function* () {
        window.open = jest.fn();
        yield help.activate(app, mainMenu);
        app.commands.execute('jupyterlab-pachyderm:open-docs');
        expect(window.open).toHaveBeenCalledWith('https://docs.pachyderm.com/latest/getting_started/');
    }));
    it('should execute the contact-support command', () => __awaiter(void 0, void 0, void 0, function* () {
        const mockDialog = showDialog;
        yield help.activate(app, mainMenu);
        app.commands.execute('jupyterlab-pachyderm:contact-support');
        expect(mockDialog).toHaveBeenCalledTimes(1);
        const { body, title } = Object.assign({}, mockDialog.mock.calls[0][0]);
        expect(title).toMatchSnapshot();
        expect(body).toMatchSnapshot();
    }));
});
