"defs.bzl contains the machinery necessary to embed postgres tools in a test that uses the snapshot library"

LINUX_LIBRARIES = [
    "//private/apt:libpq.so.5",
    "//private/apt:libldap-2.5.so.0",
    "//private/apt:liblber-2.5.so.0",
    "//private/apt:libsasl2.so.2",
]

LINUX_VARS = {
    "pgdump": "$(rlocationpath //tools/postgres/pg_dump)",
    "psql": "$(rlocationpath //tools/postgres/psql)",
    "libpq": "$(rlocationpath //private/apt:libpq.so.5)",
    "libldap": "$(rlocationpath //private/apt:libldap-2.5.so.0)",
    "liblber": "$(rlocationpath //private/apt:liblber-2.5.so.0)",
    "libsasl": "$(rlocationpath //private/apt:libsasl2.so.2)",
}

MAC_LIBRARIES = [
    "@com_enterprisedb_get_postgresql_macos//:lib/libpq.5.dylib",
    "@com_enterprisedb_get_postgresql_macos//:lib/libpq.dylib",
]

MAC_VARS = {
    "pgdump": "$(rlocationpath //tools/postgres/pg_dump)",
    "psql": "$(rlocationpath //tools/postgres/psql)",
    "libpq": "$(rlocationpath @com_enterprisedb_get_postgresql_macos//:lib/libpq.5.dylib)",
}

snapshot_x_defs = select({
    "@platforms//os:linux": LINUX_VARS,
    "@platforms//os:macos": MAC_VARS,
})

snapshot_data = [
    "//tools/postgres/pg_dump",
    "//tools/postgres/psql",
] + select({
    "@platforms//os:linux": LINUX_LIBRARIES,
    "@platforms//os:macos": MAC_LIBRARIES,
})
