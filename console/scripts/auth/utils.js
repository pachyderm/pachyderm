const {exec} = require('child_process');
const {writeFile, readFileSync} = require('fs');
const {createInterface} = require('readline');
const {parse, stringify} = require('envfile');
const {COMMAND_SEPARATOR, ENABLE_KEY_TO_CONTINUE} = require('./constants');

const prettyCommand = async (fn) => {
  console.log(`\n${COMMAND_SEPARATOR}`);
  const res = await fn();
  if (ENABLE_KEY_TO_CONTINUE) await askQuestion('press enter to continue...');
  return res;
};

module.exports = {
  executePachCommand: (command, includePachctl = true) => {
    return new Promise((res, rej) => {
      const execOut = includePachctl ? `pachctl ${command}` : command;
      console.log('\n$ ', execOut);
      exec(execOut, (err, stdout, stderr) => {
        if (err) {
          console.log('\n-----PACHCTL-OUTPUT-START------');
          !!stderr && console.log(stderr.replace(/\n$/, ''));
          console.log('-----PACHCTL-OUTPUT-END--------\n');
          rej(err);
        } else {
          console.log('\n-----PACHCTL-OUTPUT-START------');
          !!stdout && console.log(stdout.replace(/\n$/, ''));
          console.log('-----PACHCTL-OUTPUT-END--------\n');
          res(stdout);
        }
      });
    });
  },
  writeToEnv: (variable, value) => {
    const environment = readFileSync('.env.development.local', 'utf8');
    const envObject = parse(environment) || {};

    envObject[variable] = value;
    return new Promise((res, rej) => {
      writeFile('.env.development.local', stringify(envObject), (err) => {
        if (err) {
          rej(err);
        } else {
          res();
        }
      });
    });
  },
  askQuestion: (question) => {
    const readline = createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    return new Promise((res) => {
      readline.question(question + ' ', (answer) => {
        res(answer);
        readline.close();
      });
    });
  },
  prettyCommand,
  executeCommands: async (fns) => {
    for (const fn of fns) {
      await prettyCommand(fn);
    }
  },
};
