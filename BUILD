# gazelle:ignore

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")
load("@io_bazel_rules_docker//container:image.bzl", "container_image")

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
    copts = ["-Wall"],
    linkopts = ["-fPIC"],
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
    gc_linkopts = ["-linkmode=external"],
    visibility = ["//visibility:private"],
)

container_image(
    name = "k8s_fdw_image",
    base = "@postgres_10_6//image",
    files = [
          ":libgo_fdw.so",
          ":k8s_fdw.control",
          ":k8s_fdw--0.1.sql",
    ],
    # for alpine
    # symlinks = {
    #     "/usr/local/lib/postgresql/k8s_fdw.so": "/libgo_fdw.so",
    #     "/usr/local/share/postgresql/extension/k8s_fdw.control": "/k8s_fdw.control",
    #     "/usr/local/share/postgresql/extension/k8s_fdw--0.1.sql": "/k8s_fdw--0.1.sql",
    # },
    symlinks = {
        "/usr/lib/postgresql/10/lib/k8s_fdw.so": "/libgo_fdw.so",
        "/usr/share/postgresql/10/extension/k8s_fdw.control": "/k8s_fdw.control",
        "/usr/share/postgresql/10/extension/k8s_fdw--0.1.sql": "/k8s_fdw--0.1.sql",
    }
)
