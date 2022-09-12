const fetch = require('node-fetch');
const {stats} = require(`${process.cwd()}/mochawesome-report/mochawesome.json`);

const args = process.argv.slice(2);
let app = '';

if (args[0] === '--app' && args[1]) {
  app = `${args[1]} Automation Suite: `;
}

const report = [
  process.env.CIRCLE_BUILD_URL,
  new Date(stats.start).toUTCString(),
  `${app}${stats.passes}/${stats.tests} tests passing`,
  stats.failures > 0 && '<!here>',
  '--'
].filter((x) => x).join('\n');

fetch(process.env.LIVE_AUTOMATION_SLACK_URL, {
  method: 'POST',
  headers: {
    'Content-type': 'application/json'
  },
  body: JSON.stringify({
    text: report
  })
});

if (stats.failures > 0) {
  process.exit(1);
}
