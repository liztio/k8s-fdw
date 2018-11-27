cc_binary(
    name = "libgo_fdw.so",
    deps = [
        "//lib/fdw:go_fdw.cc",
        "//include/postgresql:server",
    ],
    srcs = [
        "go_fdw.c",
    ],
    copts = [
        "-Wall",
        "-Iinclude/postgresql/server",
    ],
    linkshared = True,
    visibility = ["//visibility:public"],
)
