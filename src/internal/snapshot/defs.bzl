"defs.bzl contains the machinery necessary to embed postgres tools in a test that uses the snapshot library"

LINUX_LIBRARIES = [
    "//private/apt:libpq.so.5",
    "//private/apt:libldap-2.5.so.0",
    "//private/apt:liblber-2.5.so.0",
    "//private/apt:libsasl2.so.2",
]

us = "github.com/pachyderm/pachyderm/v2/src/internal/snapshot."

LINUX_VARS = {
    us + "pgdump": "$(rlocationpath //tools/postgres/pg_dump)",
    us + "psql": "$(rlocationpath //tools/postgres/psql)",
    us + "libpq": "$(rlocationpath //private/apt:libpq.so.5)",
    us + "libldap": "$(rlocationpath //private/apt:libldap-2.5.so.0)",
    us + "liblber": "$(rlocationpath //private/apt:liblber-2.5.so.0)",
    us + "libsasl": "$(rlocationpath //private/apt:libsasl2.so.2)",
}

MAC_LIBRARIES = [
    "@com_enterprisedb_get_postgresql_macos//:lib/libpq.5.dylib",
    "@com_enterprisedb_get_postgresql_macos//:lib/libpq.dylib",
]

MAC_VARS = {
    us + "pgdump": "$(rlocationpath //tools/postgres/pg_dump)",
    us + "psql": "$(rlocationpath //tools/postgres/psql)",
    us + "libpq": "$(rlocationpath @com_enterprisedb_get_postgresql_macos//:lib/libpq.5.dylib)",
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
