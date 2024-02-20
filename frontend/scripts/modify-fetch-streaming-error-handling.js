/* eslint-disable @typescript-eslint/no-var-requires */

const {Project} = require('ts-morph');
const {SemicolonPreference} = require('typescript');

const modifyFunction = (
  func,
  functionName,
  originalFunction,
  modifiedFunction,
) => {
  if (func.getName() === functionName) {
    if (func.getBodyText() !== originalFunction) {
      console.log(
        `WARNING: '${functionName}' has been changed by maintainer. Our modifications may need to be updated`,
      );
    }

    func.setBodyText(modifiedFunction);
    return true;
  }
  return false;
};

const modifyFetchStreamingRequest = (func) => {
  const originalFunction =
    'const {pathPrefix, ...req} = init || {}\n' +
    // eslint-disable-next-line no-template-curly-in-string
    '  const url = pathPrefix ?`${pathPrefix}${path}` : path\n' +
    '  const result = await fetch(url, req)\n' +
    '  // needs to use the .ok to check the status of HTTP status code\n' +
    '  // http other than 200 will not throw an error, instead the .ok will become false.\n' +
    '  // see https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch#\n' +
    '  if (!result.ok) {\n' +
    '    const resp = await result.json()\n' +
    '    const errMsg = resp.error && resp.error.message ? resp.error.message : ""\n' +
    '    throw new Error(errMsg)\n' +
    '  }\n' +
    '\n' +
    '  if (!result.body) {\n' +
    '    throw new Error("response doesnt have a body")\n' +
    '  }\n' +
    '\n' +
    '  await result.body\n' +
    '    .pipeThrough(new TextDecoderStream())\n' +
    '    .pipeThrough<R>(getNewLineDelimitedJSONDecodingStream<R>())\n' +
    '    .pipeTo(getNotifyEntityArrivalSink((e: R) => {\n' +
    '      if (callback) {\n' +
    '        callback(e)\n' +
    '      }\n' +
    '    }))\n' +
    '\n' +
    '  // wait for the streaming to finish and return the success respond\n' +
    '  return';

  const modifiedFunction = `const {pathPrefix, ...req} = init || {};
const url = pathPrefix ? \`\${pathPrefix}\${path}\` : path;
const result = await fetch(url, req);
// needs to use the .ok to check the status of HTTP status code
// http other than 200 will not throw an error, instead the .ok will become false.
// see https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch#
if (!result.ok) {
const contentType = result.headers.get('Content-Type');
let resp;
if (contentType === 'application/json'){
  const resp = await result.json();
  if (typeof resp === 'object' && 'error' in resp) {
    throw resp.error;
  }
} else if (contentType === 'text/plain') {
  resp = await result.text();
  throw new Error(resp);  
}
throw resp;
}

if (!result.body) {
throw new Error('response doesnt have a body');
}

await result.body
.pipeThrough(new TextDecoderStream())
.pipeThrough<R>(getNewLineDelimitedJSONDecodingStream<R>())
.pipeTo(
  getNotifyEntityArrivalSink((e: R) => {
    if (e.error) {
      throw e.error;
    }
    if (callback) {
      callback(e.result);
    }
  }),
);

// wait for the streaming to finish and return the success respond
return;`;

  return modifyFunction(
    func,
    'fetchStreamingRequest',
    originalFunction,
    modifiedFunction,
  );
};

const modifyFetchReq = (func) => {
  const originalFunction =
    'const {pathPrefix, ...req} = init || {}\n\n' +
    // eslint-disable-next-line no-template-curly-in-string
    '  const url = pathPrefix ? `${pathPrefix}${path}` : path\n\n' +
    '  return fetch(url, req).then(r => r.json().then((body: O) => {\n' +
    '    if (!r.ok) { throw body; }\n' +
    '    return body;\n' +
    '  })) as Promise<O>';

  const modifiedFunction = `const { pathPrefix, ...req } = init || {};

  const url = pathPrefix ? \`\${pathPrefix}\${path}\` : path;

  return fetch(url, req).then((r) => {
    const contentType = r.headers.get('Content-Type');
    if (contentType === 'application/json') {
      return r.json().then((body: O) => {
        if (!r.ok) {
          throw body;
        }
        return body;
      });
    } else if (contentType === 'text/plain') {
      return r.text().then((body: O) => {
        if (!r.ok) {
          throw new Error(body);
        }
        return body;
      });
    } else {
      if (!r.ok) {
        throw body;
      }
      return body;
    }
  }) as Promise<O>;`;

  return modifyFunction(func, 'fetchReq', originalFunction, modifiedFunction);
};

const modifyGetNewLineDelimitedJSONDecodingStream = (func) => {
  const originalFunction =
    'return new TransformStream({\n' +
    '    start(controller: JSONStringStreamController<T>) {\n' +
    "      controller.buf = ''\n" +
    '      controller.pos = 0\n' +
    '    },\n' +
    '\n' +
    '    transform(chunk: string, controller: JSONStringStreamController<T>) {\n' +
    '      if (controller.buf === undefined) {\n' +
    "        controller.buf = ''\n" +
    '      }\n' +
    '      if (controller.pos === undefined) {\n' +
    '        controller.pos = 0\n' +
    '      }\n' +
    '      controller.buf += chunk\n' +
    '      while (controller.pos < controller.buf.length) {\n' +
    "        if (controller.buf[controller.pos] === '\\n') {\n" +
    '          const line = controller.buf.substring(0, controller.pos)\n' +
    '          const response = JSON.parse(line)\n' +
    '          controller.enqueue(response.result)\n' +
    '          controller.buf = controller.buf.substring(controller.pos + 1)\n' +
    '          controller.pos = 0\n' +
    '        } else {\n' +
    '          ++controller.pos\n' +
    '        }\n' +
    '      }\n' +
    '    }\n' +
    '  })';

  const modifiedFunction = `return new TransformStream({
    start(controller: JSONStringStreamController<T>) {
        controller.buf = '';
        controller.pos = 0;
    },

    transform(chunk: string, controller: JSONStringStreamController<T>) {
        if (controller.buf === undefined) {
            controller.buf = '';
        }
        if (controller.pos === undefined) {
            controller.pos = 0;
        }
        controller.buf += chunk;
        while (controller.pos < controller.buf.length) {
            if (controller.buf[controller.pos] === '\\n') {
                const line = controller.buf.substring(0, controller.pos);
                const response = JSON.parse(line);
                controller.enqueue(response);
                controller.buf = controller.buf.substring(controller.pos + 1);
                controller.pos = 0;
            } else {
                ++controller.pos;
            }
        }
    }
});`;

  return modifyFunction(
    func,
    'getNewLineDelimitedJSONDecodingStream',
    originalFunction,
    modifiedFunction,
  );
};

const modifyFetchStreamingErrorHandling = (
  sourceFilePattern = './src/generated/proto/fetch.pb.ts',
) => {
  const project = new Project({});
  let modifiedFunctionsCount = 0;

  project.addSourceFilesAtPaths(sourceFilePattern);
  project.getSourceFiles().forEach((sourceFile) => {
    sourceFile.getFunctions().forEach((func) => {
      if (modifyFetchStreamingRequest(func)) {
        modifiedFunctionsCount += 1;
      } else if (modifyGetNewLineDelimitedJSONDecodingStream(func)) {
        modifiedFunctionsCount += 1;
      } else if (modifyFetchReq(func)) {
        modifiedFunctionsCount += 1;
      }
    });

    sourceFile.formatText({
      semicolons: SemicolonPreference.Insert,
      tabSize: 2,
      convertTabsToSpaces: true,
    });
  });

  project.saveSync();

  if (modifiedFunctionsCount !== 3) {
    console.log(
      `ERROR: Modified ${modifiedFunctionsCount} function(s). Wanted to modify exactly 3.`,
    );
  } else {
    console.log(
      `SUCCESS: Modified functions 'fetchStreamingRequest', 'fetchReq', and 'getNewLineDelimitedJSONDecodingStream' in fetch.pb.ts`,
    );
  }
};

module.exports = {
  modifyFetchStreamingErrorHandling,
};
