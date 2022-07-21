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
import { load, track } from 'rudder-sdk-js';
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
    let app;
    beforeEach(() => {
        app = new JupyterLab();
    });
    it('should have the correct configuration', () => {
        expect(telemetry.id).toBe('jupyterlab-pachyderm:telemetry');
        expect(telemetry.autoStart).toBe(true);
    });
    it('should initialize logging', () => __awaiter(void 0, void 0, void 0, function* () {
        yield telemetry.activate(app);
        expect(load).toHaveBeenCalledWith('20C6D2xFLRmyFTqtvYDEgNfwcRG', 'https://pachyderm-dataplane.rudderstack.com');
    }));
    it('should track app commands', () => __awaiter(void 0, void 0, void 0, function* () {
        yield telemetry.activate(app);
        app.commands.addCommand('fake:command', { execute: jest.fn() });
        yield app.commands.execute('fake:command', { fake: 'args' });
        expect(track).toHaveBeenCalledWith('command', {
            id: 'fake:command',
            args: { fake: 'args' },
        });
    }));
    it('should track notebook commands with a prompt', () => __awaiter(void 0, void 0, void 0, function* () {
        yield telemetry.activate(app);
        expect(track).toHaveBeenCalledWith('command', {
            id: 'notebook:action:executed',
            args: {
                action: 'pachctl version',
            },
        });
    }));
    it('should track notebook commands without a prompt', () => __awaiter(void 0, void 0, void 0, function* () {
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
        yield telemetry.activate(app);
        expect(track).toHaveBeenCalledWith('command', {
            id: 'notebook:action:executed',
            args: {
                action: 'Hello Telemetry',
            },
        });
    }));
});
