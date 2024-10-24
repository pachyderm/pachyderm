# This is modified from https://github.com/bazel-contrib/bazel-lib/blob/main/lib/private/modify_mtree.awk
{
    if (strip_prefix != "") {
        if ($1 == strip_prefix) {
            # this line declares the directory which is now the root. It may be discarded.
            next;
        } else if (index($1, strip_prefix) == 1) {
            # this line starts with the strip_prefix
            sub("^" strip_prefix "/", "");

            # NOTE: The mtree format treats file paths without slashes as "relative" entries.
            #       If a relative entry is a directory, then it will "change directory" to that
            #       directory, and any subsequent "relative" entries will be created inside that
            #       directory. This causes issues when there is a top-level directory that is
            #       followed by a top-level file, as the file will be created inside the directory.
            #       To avoid this, we append a slash to the directory path to make it a "full" entry.
            components = split($1, _, "/");
            if ($0 ~ /type=dir/ && components == 1) {
                $1 = $1 "/";
            }
        } else {
            # this line declares some path under a parent directory, which will be discarded
            next;
        }
    }

    sub(/uid=[0-9\.]+/, "uid=65532");
    sub(/gid=[0-9\.]+/, "gid=65532");

    if (package_dir != "") {
        sub(/^/, package_dir "/");
    }

    if ($0 ~ /init.runfiles\//) {
        next;
    }
    print;
}
