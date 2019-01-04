load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "fdw.go",
        "go_fdw.go",
    ],
    cdeps = [
        "//include/postgresql:internal",  # keep
        "//include/postgresql:server",  # keep
    ],
    cgo = True,
    copts = ["-Iinclude/postgresql/server -Iinclude/postgresql/internal"],  # keep
    clinkopts = ["-Wl,-unresolved-symbols=ignore-all"], # keep
    importpath = "github.com/liztio/k8s-fdw/lib/fdw",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/client:go_default_library",
        "//vendor/github.com/pkg/errors:go_default_library",
    ],
)

go_binary(
    name = "go_fdw",
    embed = [":go_default_library"],
    linkmode = "c-archive",  # keep
    visibility = ["//visibility:public"],
)
