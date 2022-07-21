var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { fileIcon } from '../../utils/icons';
const EXAMPLES_CATAGORY = 'Pachyderm Examples';
const examples = [
    {
        path: 'examples/Intro to Pachyderm Tutorial.ipynb',
        title: 'Intro to Pachyderm',
        command: 'jupyterlab-pachyderm:open-example-intro',
        rank: 1,
    },
    {
        path: 'examples/Mounting Data Repos in Notebooks.ipynb',
        title: 'Mounting Data Repos',
        command: 'jupyterlab-pachyderm:open-example-mount',
        rank: 2,
    },
];
const doesPathExist = (path) => __awaiter(void 0, void 0, void 0, function* () {
    try {
        const response = yield fetch(`/api/contents/${path}`);
        return response.status === 200;
    }
    catch (_a) {
        return false;
    }
});
const addExample = (example, app, launcher) => __awaiter(void 0, void 0, void 0, function* () {
    if (yield doesPathExist(example.path)) {
        app.commands.addCommand(example.command, {
            icon: fileIcon,
            label: example.title,
            execute: () => {
                return app.commands.execute('docmanager:open', {
                    path: example.path,
                });
            },
        });
        launcher.add({
            command: example.command,
            category: EXAMPLES_CATAGORY,
            rank: example.rank,
        });
    }
});
export const init = (app, launcher) => __awaiter(void 0, void 0, void 0, function* () {
    yield Promise.all(examples.map((example) => {
        return addExample(example, app, launcher);
    }));
});
