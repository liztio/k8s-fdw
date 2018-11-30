cc_binary(
    name = "libk8s_fdw.so",
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
    visibility = ["//visibility:private"],
)

# cc_binary required to start with lib, strip that out
genrule(
    name = "strip_lib",
    srcs = [":libk8s_fdw.so"],
    outs = ["k8s_fdw.so"],
    cmd = "cp $< $@",
    executable = True,
    visibility = ["//visibility:public"],
)
