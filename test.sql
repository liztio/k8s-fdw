CREATE EXTENSION IF NOT EXISTS k8s_fdw;

CREATE SERVER IF NOT EXISTS kind
  FOREIGN DATA WRAPPER k8s_fdw
  OPTIONS (kubeconfig '/kubeconfig');

CREATE FOREIGN TABLE IF NOT EXISTS pods (
  name varchar OPTIONS (alias 'metadata.name')
, namespace varchar OPTIONS (alias 'metadata.namespace')
--, container varchar OPTIONS (alias '{.spec.containers[0].image}')
)
  SERVER kind
  OPTIONS (
    namespace 'kube-system',
    apiVersion 'v1',
    kind 'Pod'
  );

SELECT * FROM PODS;
