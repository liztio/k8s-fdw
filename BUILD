load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/example/project
gazelle(name = "gazelle")

go_library(
    name = "go_default_library",
    srcs = [
        "fdw.go",
        "go_fdw.c",
        "go_fdw.go",
        "//include/postgresql:internal",  # keep
        "//include/postgresql:server",  # keep
    ],
    cgo = True,
    copts = ["-Iinclude/postgresql/server -Iinclude/postgresql/internal"],
    importpath = "github.com/example/project",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "k8s-fdw.so",
    embed = [":go_default_library"],
    linkmode = "c-archive",
    visibility = ["//visibility:public"],
)