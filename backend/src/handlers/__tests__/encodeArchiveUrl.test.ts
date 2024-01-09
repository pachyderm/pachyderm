import {createRequest, createResponse} from 'node-mocks-http';

import {encodeArchiveUrl} from '../handleEncodeArchiveUrl';

describe('EncodeArchiceUrl', () => {
  it('should correctly encode url from request', async () => {
    const request = createRequest({
      body: {
        projectId: 'default',
        repoId: 'images',
        commitId: '5b31d378267c4e7a8f260edbd7811c3d',
        paths: ['/AT-AT.png', '/cats/', '/json_nested_arrays.json'],
      },
    });

    const response = createResponse();
    await encodeArchiveUrl(request, response);

    expect(JSON.parse(response._getData())).toEqual({
      url: '/proxyForward/archive/ASi1L_0guvUCAHKFFBpwSR1wND_wozPYnfTk738G-KbSdtipquqdH478UX6xSbKOe367c-QP83OhIT1RGgM0Zgic4DlL-XEHFIprtQVQFVqa-KQ3KHzLRRjMPCg2NgcCAGsV6dVHCgo.zip',
    });
  });
});
