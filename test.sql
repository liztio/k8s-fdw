CREATE EXTENSION IF NOT EXISTS k8s_fdw;

CREATE SERVER IF NOT EXISTS kind
  FOREIGN DATA WRAPPER k8s_fdw
  OPTIONS (kubeconfig '/kubeconfig');

CREATE FOREIGN TABLE IF NOT EXISTS pods (
  name      text OPTIONS (alias 'metadata.name')
, namespace text OPTIONS (alias 'metadata.namespace')
, container text OPTIONS (alias '{.spec.containers[0].image}')
-- , phase     text OPTIONS (alias 'status.phase')
-- , reason    text OPTIONS (alias 'status.reason')
)
  SERVER kind
  OPTIONS (
    namespace 'kube-system'
  , apiVersion 'v1'
  , kind 'Pod'
  );

SELECT * FROM PODS;
