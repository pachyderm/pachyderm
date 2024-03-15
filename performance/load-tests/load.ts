import {BROWSER_COUNT, TEST_RUNTIME} from './src/constants';
import {
  enableShutDownFlag,
  runJourney,
  setupContexts,
  startBrowsers,
  tearDown,
} from './src/browser';
import {TestIsFinished} from './src/utils';
import {Browser, BrowserContext} from 'playwright';
import {statsLogger} from './src/logger';

const testTimeout = async (ms: number) =>
  new Promise((_, reject) =>
    setTimeout(() => {
      console.log('The test has finished!');
      reject(new TestIsFinished('Test timed out.'));
    }, ms),
  );

// Shut down stuff
// I am not sure if this is needed.
let browsers: Browser[] = [];
let contexts: BrowserContext[] = [];

const safelyShutdownContextAndBrowsers = async () => {
  enableShutDownFlag();
  await Promise.all(contexts.map(tearDown));
  await Promise.all(browsers.map(tearDown));
};

process.on('SIGINT', async () => {
  console.log('Caught interrupt signal, closing browsers');
  statsLogger.playback();
  await safelyShutdownContextAndBrowsers();
  process.exit();
});
// Shut down stuff ^

const loadTest = async () => {
  try {
    // Setup
    browsers = await startBrowsers(BROWSER_COUNT);
    contexts = await setupContexts(browsers);

    // This will run all journeys in parallel as well as the test timeout fn.
    // Test timeout will fire after x minutes and end the test!
    const runJourneysForever = contexts.map((context, index) =>
      runJourney(context, index + 1),
    );
    await Promise.race([testTimeout(TEST_RUNTIME), ...runJourneysForever]);
    await safelyShutdownContextAndBrowsers();
  } catch (e) {
    if (e instanceof TestIsFinished) {
      // Good!
      console.log('The test has finished!');
    } else {
      // Very bad!
      console.log('Uncaught error!');
      console.error(e);
      throw e;
    }
  } finally {
    statsLogger.playback();
  }

  process.exit();
};

loadTest();
