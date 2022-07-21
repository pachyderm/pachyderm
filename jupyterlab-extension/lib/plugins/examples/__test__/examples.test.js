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
import { LauncherModel } from '@jupyterlab/launcher';
import examples from '../';
describe('examples plugin', () => {
    let app;
    let launcher;
    beforeEach(() => {
        app = new JupyterLab();
        launcher = new LauncherModel();
    });
    it('should add example if the file is found', () => __awaiter(void 0, void 0, void 0, function* () {
        fetchMock.mockResponse('');
        yield examples.activate(app, launcher);
        const appCommands = app.commands.listCommands();
        expect(appCommands.length).toEqual(2);
        expect(appCommands[0]).toEqual('jupyterlab-pachyderm:open-example-intro');
        expect(appCommands[1]).toEqual('jupyterlab-pachyderm:open-example-mount');
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
    }));
    it('should not add example if the file is not found', () => __awaiter(void 0, void 0, void 0, function* () {
        fetchMock.mockReject();
        yield examples.activate(app, launcher);
        expect(app.commands.listCommands().length).toEqual(0);
        expect(launcher.items().next()).toBeUndefined();
    }));
});
