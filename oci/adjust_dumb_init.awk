{
    sub(/^file\//, "");
    sub(/uid=[0-9\.]+/, "uid=65532");
    sub(/gid=[0-9\.]+/, "gid=65532");

    if (package_dir != "") {
        sub(/^/, package_dir "/");
    }

    print;
}
