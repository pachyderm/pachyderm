import {LanguageSupport, StreamLanguage} from '@codemirror/language';
import * as yamlMode from '@codemirror/legacy-modes/mode/yaml';

const yaml = () => new LanguageSupport(StreamLanguage.define(yamlMode.yaml));

export default yaml;
