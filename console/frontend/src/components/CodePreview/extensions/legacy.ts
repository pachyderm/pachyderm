import {
  LanguageSupport,
  StreamLanguage,
  StreamParser,
} from '@codemirror/language';
import * as dockerFileMode from '@codemirror/legacy-modes/mode/dockerfile';
import * as goMode from '@codemirror/legacy-modes/mode/go';
import * as juliaMode from '@codemirror/legacy-modes/mode/julia';
import * as protobufMode from '@codemirror/legacy-modes/mode/protobuf';
import * as rMode from '@codemirror/legacy-modes/mode/r';
import * as rubyMode from '@codemirror/legacy-modes/mode/ruby';
import * as shellMode from '@codemirror/legacy-modes/mode/shell';
import * as yamlMode from '@codemirror/legacy-modes/mode/yaml';

const createLegacyExtension = (mode: StreamParser<unknown>) =>
  new LanguageSupport(StreamLanguage.define(mode));

export const docker = () => createLegacyExtension(dockerFileMode.dockerFile);
export const go = () => createLegacyExtension(goMode.go);
export const julia = () => createLegacyExtension(juliaMode.julia);
export const protobuf = () => createLegacyExtension(protobufMode.protobuf);
export const r = () => createLegacyExtension(rMode.r);
export const ruby = () => createLegacyExtension(rubyMode.ruby);
export const shell = () => createLegacyExtension(shellMode.shell);
export const yaml = () => createLegacyExtension(yamlMode.yaml);
