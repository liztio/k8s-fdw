# gazelle:ignore

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/example/project
gazelle(name = "gazelle")

cc_binary(
    name = "libgo_fdw.so",
    deps = [
      ":go_fdw.cc",
      "//include/postgresql:server",
    ],
    srcs = [
        ":go_fdw.c_hdrs",
        "go_fdw.c",
    ],
    copts = [
           "-Wall"
    ],
    linkshared = True,
    includes = ["include/postgresql/server"],
)

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
    copts = ["-Iinclude/postgresql/server -Iinclude/postgresql/internal"],
    importpath = "github.com/example/project",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "go_fdw",
    embed = [":go_default_library"],
    linkmode = "c-archive",
    visibility = ["//visibility:private"],
)