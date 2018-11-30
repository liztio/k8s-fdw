CREATE EXTENSION IF NOT EXISTS k8s_fdw;

CREATE SERVER IF NOT EXISTS kind
  FOREIGN DATA WRAPPER k8s_fdw
  OPTIONS (kubeconfig '/kubeconfig');

CREATE FOREIGN TABLE IF NOT EXISTS pods (
  name      text OPTIONS (alias 'metadata.name')
, namespace text OPTIONS (alias 'metadata.namespace')
, container text OPTIONS (alias '{.spec.containers[0].image}')
, labels   jsonb OPTIONS (alias 'metadata.labels')
)
  SERVER kind
  OPTIONS (
    namespace 'kube-system'
  , apiVersion 'v1'
  , kind 'Pod'
  );

CREATE FOREIGN TABLE IF NOT EXISTS replica_sets (
  name      text OPTIONS (alias 'metadata.name')
, replicas  bigint OPTIONS (alias 'status.replicas')
, available bigint OPTIONS (alias 'status.availableReplicas')
)
  SERVER kind
  OPTIONS (
    namespace 'kube-system'
  , apiVersion 'apps/v1'
  , kind 'ReplicaSet'
  );

CREATE FOREIGN TABLE IF NOT EXISTS deployments (
  name      text OPTIONS (alias 'metadata.name')
, replicas  bigint OPTIONS (alias 'status.replicas')
, available bigint OPTIONS (alias 'status.availableReplicas')
)
  SERVER kind
  OPTIONS (
    namespace 'kube-system'
  , apiVersion 'apps/v1'
  , kind 'Deployment'
  );

SELECT * FROM replica_sets;

SELECT "deployments"."name" AS deployment_name
     , "replica_sets"."name" as replica_name
     , "pods"."name" AS pod_name
  FROM deployments
  JOIN replica_sets on "replica_sets"."name" LIKE "deployments"."name" || '-%'
  JOIN pods on "pods"."name" LIKE "replica_sets"."name" || '-%';
