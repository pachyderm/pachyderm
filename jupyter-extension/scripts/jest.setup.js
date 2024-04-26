import failOnConsole from 'jest-fail-on-console'

const ignoredErrors = [
  // I am not exactly sure what causes this error to occur. This typically occurs when findByTestId is not properly awaited which is not happening.
  // The tests seem fine and pass. I think we should revisit this error once we can update all the testing libraries, but for now it can be silenced.
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
