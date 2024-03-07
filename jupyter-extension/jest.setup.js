const originalError = console.error.bind(console.error)
const ignoredErrors = [
  // I am not exactly sure what causes this error to occur. This typically occurs when findByTestId is not properly awaited which is not happening.
  // The tests seem fine and pass. I think we should revisit this error once we can update all the testing libraries, but for now it can be silenced.
  (msg, optionalParams) => msg.includes('Warning: An update to %s inside a test was not wrapped in act(...)') && optionalParams[0] === 'Datum'
]
const originalWarn = console.warn.bind(console.warn)
const ignoredWarnings = [
    // For some reason JupyterLab's own LabIcons from @jupyterlab/ui-components complain they are malformed. This issue is fixed in future versions
    // of the library, but for now can be safely ignored.
    (msg, optionalParams) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:caret-up'),
    (msg, optionalParams) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:caret-down'),
    (msg, optionalParams) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:caret-right'),
    (msg, optionalParams) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:caret-left'),
    (msg, optionalParams) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:close'),
    (msg, optionalParams) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:folder'),
    (msg, optionalParams) => msg.includes('SVG HTML was malformed for LabIcon instance.') && msg.includes('name: ui-components:ellipses')
]


beforeAll(() => {
  console.error = (msg, ...optionalParams) => {
    for (const ignoredError of ignoredErrors) {
        if (ignoredError(msg, optionalParams)) {
            return;
        }
    }
    originalError(msg)
  }
  console.warn = (msg, ...optionalParams) => {
    for (const ignoredWarning of ignoredWarnings) {
        if (ignoredWarning(msg, optionalParams)) {
            return;
        }
    }
    originalWarn(msg)
  }
})
afterAll(() => {
  console.warn = originalWarn
  console.error = originalError
})
