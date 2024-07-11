// NOTE: This is temporarily unused see `jest.config.js` for where it is commented out.

import failOnConsole from 'jest-fail-on-console'

const ignoredErrors = [
  // This error is caused because we are running two different versions of react in our application. JupyterLab depends on a higher
  // version of react than our extension does. React supports this behavior by design, but this is something we should eventually address by
  // updating our version of React. Two different versions of React can cause some unexpected problems given the difference between the two is a major version and
  // the ability to run two different versions of react is meant primarily for entirely separate areas of an application not tightly coupled like JupyterLab and our extension.
  (msg) => msg.includes('Warning: An update to Datum inside a test was not wrapped in act(...)')
]
const ignoredWarnings = [
  // For some reason JupyterLab's own LabIcons from @jupyterlab/ui-components complain they are malformed. This issue is fixed in future versions
  // of the library, but for now can be safely ignored.
  (msg) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:caret-up'),
  (msg) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:caret-down'),
  (msg) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:caret-right'),
  (msg) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:caret-left'),
  (msg) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:close'),
  (msg) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:folder'),
  (msg) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:ellipses')
]

failOnConsole({
  shouldFailOnAssert: true,
  shouldFailOnDebug: true,
  shouldFailOnError: true,
  shouldFailOnInfo: true,
  shouldFailOnLog: true,
  shouldFailOnWarn: true,
  silenceMessage: (message, methodName) => {
    if (methodName === 'error') {
      for (const ignoredError of ignoredErrors) {
        if (ignoredError(message)) {
            return true;
        }
      }
    }

    if (methodName === 'warn') {
      for (const ignoredWarning of ignoredWarnings) {
        if (ignoredWarning(message)) {
            return true;
        }
      }
    }

    return false;
  },
})
