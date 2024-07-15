# Proto Generation

This requires a separate package.json, as the proto generation libraries don't include linux arm build, and thus the build fails on circle arm. It does work locally on an m1 mac though.

## Important Note

After generating the protos, make sure to remove the node_modules from this folder. Apparently the parent application mistakenly references node_modules from this directory, leading to a potential app malfunction.

### Error Message to Look Out For

```bash
[{"extensions": {"code": "INTERNAL_SERVER_ERROR", "details": "Expected argument of type google.protobuf.Empty", "grpcCode": 13, "metadata": {"content-type": ["application/grpc+proto"], "date": ["Tue, 01 Aug 2023 00:19:41 GMT"]}}, "locations": [{"column": 3, "line": 2}], "message": "Something went wrong", "path": ["finishCommit"]}]
```

Should you encounter the error above, the first step is to delete the proto node_modules.
