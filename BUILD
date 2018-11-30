load("@bazel_gazelle//:def.bzl", "gazelle")
load("@io_bazel_rules_docker//container:image.bzl", "container_image")

# gazelle:prefix github.com/liztio/k8s-fdw
gazelle(
    name = "gazelle",
)

container_image(
    name = "k8s_fdw_image",
    base = "@postgres_10_6//image",
    files = [
        "//src:libgo_fdw.so",
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
    },
    stamp = True,
)
