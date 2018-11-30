load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")
load("@bazel_gazelle//:def.bzl", "gazelle")
load("@io_bazel_rules_docker//container:image.bzl", "container_image")
load("@io_bazel_rules_docker//container:push.bzl" , "container_push")

# gazelle:prefix github.com/liztio/k8s-fdw
gazelle(
    name = "gazelle",
)

filegroup(
    name ="artifacts",
    srcs = [
        "//src:k8s_fdw.so",
        ":k8s_fdw.control",
        ":k8s_fdw--0.1.sql",
    ],
)

container_image(
    name = "k8s_fdw_image",
    base = "@postgres_10_6//image",
    files = [":artifacts"],
    # for alpine
    # symlinks = {
    #     "/usr/local/lib/postgresql/k8s_fdw.so": "/libgo_fdw.so",
    #     "/usr/local/share/postgresql/extension/k8s_fdw.control": "/k8s_fdw.control",
    #     "/usr/local/share/postgresql/extension/k8s_fdw--0.1.sql": "/k8s_fdw--0.1.sql",
    # },
    symlinks = {
        "/usr/lib/postgresql/10/lib/k8s_fdw.so": "/k8s_fdw.so",
        "/usr/share/postgresql/10/extension/k8s_fdw.control": "/k8s_fdw.control",
        "/usr/share/postgresql/10/extension/k8s_fdw--0.1.sql": "/k8s_fdw--0.1.sql",
    },
    stamp = True,
)

container_push(
    name = "push_latest",
    format = "Docker",
    image = ":k8s_fdw_image",
    registry = "index.docker.io",
    repository = "liztio/k8s_fdw",
    tag = "latest",
)

pkg_tar(
    name = "k8s_fdw_release",
    extension = "tar.gz",
    srcs = [":artifacts"],
    modes = {
        "k8s_fdw.so": "0755",
        "k8s_fdw.control": "0644",
        "k8s_fdw--0.1.sql": "0644",
    },
)

genrule(
    name = "release_checksum",
    srcs = [":k8s_fdw_release"],
    outs = ["k8s_fdw_release_checksums.txt"],
    # Some shenanigans to strip all the bazel-bin/ out of the checksums
    cmd = "(cd $$(dirname $<) && sha256sum $$(basename $<)) > $@",
)

filegroup(
    name = "release",
    srcs = [
        ":k8s_fdw_release",
        ":release_checksum",
    ],
)
