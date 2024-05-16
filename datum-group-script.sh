pachctl create repo input;
pachctl put file input@master:/Aquatroll/2024/01/01/12345/data -f ent.txt;
pachctl put file input@master:/Aquatroll/2024/01/01/12346/data -f ent.txt;
pachctl put file input@master:/Leveltroll/2024/01/01/12345/data -f ent.txt;
pachctl put file input@master:/Aquatroll/2024/01/01/12345/flags -f ent.txt;
pachctl put file input@master:/Aquatroll/2024/04/02/1234/data -f ent.txt;
pachctl put file input@master:/Aquatroll/2024/04/03/1235/data -f ent.txt;
pachctl list datum -f test-pipe.json
