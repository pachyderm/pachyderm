import {createRequest, createResponse} from 'node-mocks-http';

import {encodeArchiveUrl} from '../handleEncodeArchiveUrl';

describe('EncodeArchiceUrl', () => {
  it('should correctly encode url from request', async () => {
    const request = createRequest({
      body: {
        projectId: 'default',
        repoId: 'images',
        commitId: '5b31d378267c4e7a8f260edbd7811c3d',
        paths: ['/fruit.png', '/cats/', '/json_nested_arrays.json'],
      },
    });

    const response = createResponse();
    await encodeArchiveUrl(request, response);

    expect(JSON.parse(response._getData())).toEqual({
      url: '/proxyForward/archive/ASi1L_0guuUCAHIFFBmAhw6AP1rgZ5QcddIBBnqHazbQLoCBGY0YbXhBPKEx6lp-515tePEaQR4e0QGKv7ZCWNSMUblsOZCOU5osaJmOSrQGsUHB-ekBljcc6ngDAgB7FZ1TRwoK.zip',
    });
  });
});
