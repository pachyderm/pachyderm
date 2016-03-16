1 - finish ripping out pps into /client
2 - make sure localtest pass
3 - push to CI and make sure integration tests pass
4 - rip out pfsutil and ppsutil into /client repsectively
5 - pass CI
6 - rename /src/pps to /src/server/pps ... and for pfs
7 - pass CI
--> 8 - apply vendoring
9 - pass CI
10 - update sandbox
11 - pass CI and deploy
12 - sandbox connect to pachd in production