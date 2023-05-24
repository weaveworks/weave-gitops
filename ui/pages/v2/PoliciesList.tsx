import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";
import { PolicyTable } from "../../components/Policies/PolicyList/PolicyTable";
import { useListListPolicies } from "../../hooks/Policies";

type Props = {
  className?: string;
};

function PoliciesList({ className }: Props) {
  const { data, isLoading, error } = useListListPolicies({});

  // const isLoading: boolean = false;
  // const error = null;
  // const data = {
  //   policies: [
  //     {
  //       category: "weave.categories.reliability",
  //       clusterName: "management",
  //       code: 'package weave.advisor.pods.replica_count\n\nimport future.keywords.in\n\nmin_replica_count := input.parameters.replica_count\nexclude_namespaces := input.parameters.exclude_namespaces\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\ncontroller_input := input.review.object\n\nviolation[result] {\n\tisExcludedNamespace == false\n    not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n\tnot replicas >= min_replica_count\n\tresult = {\n\t\t"issue detected": true,\n\t\t"msg": sprintf("Replica count must be greater than or equal to \'%v\'; found \'%v\'.", [min_replica_count, replicas]),\n\t\t"violating_key": violating_key,\n\t\t"recommended_value": min_replica_count,\n\t}\n}\n\nreplicas := controller_input.spec.replicas {\n\tcontroller_input.kind in {"Deployment", "StatefulSet", "ReplicaSet", "ReplicationController"}\n} else := controller_input.spec.minReplicas {\n\tcontroller_input.kind == "HorizontalPodAutoscaler"\n}\n\nviolating_key := "spec.replicas" {\n\tcontroller_input.kind in {"Deployment", "StatefulSet", "ReplicaSet", "ReplicationController"}\n} else := "spec.minReplicas" {\n\tcontroller_input.kind == "HorizontalPodAutoscaler"\n}\n\nisExcludedNamespace = true {\n\tcontroller_input.metadata.namespace\n\tcontroller_input.metadata.namespace in exclude_namespaces\n} else = false',
  //       createdAt: "2023-02-28T10:02:09Z",
  //       description:
  //         "Use this Policy to to check the replica count of your workloads. The value set in the Policy is greater than or equal to the amount desired, so if the replica count is lower than what is specified, the Policy will be in violation. \n",
  //       gitCommit: "",
  //       howToSolve:
  //         "The replica count should be a value equal or greater than what is set in the Policy.\n```\nspec:\n  replicas: <replica_count>\n```\nhttps://kubernetes.io/docs/concepts/workloads/controllers/deployment/#scaling-a-deployment\n",
  //       id: "weave.policies.containers-minimum-replica-count",
  //       modes: ["audit", "admission"],
  //       name: "Containers Minimum Replica Count",
  //       parameters: [
  //         { name: "replica_count", type: "integer", required: true },
  //         {
  //           name: "exclude_namespaces",
  //           type: "array",
  //           value: null,
  //           required: false,
  //         },
  //         {
  //           name: "exclude_label_key",
  //           type: "string",
  //           value: null,
  //           required: false,
  //         },
  //         {
  //           name: "exclude_label_value",
  //           type: "string",
  //           value: null,
  //           required: false,
  //         },
  //       ],
  //       severity: "medium",
  //       standards: [
  //         {
  //           id: "weave.standards.soc2-type-i",
  //           controls: ["weave.controls.soc2-type-i.2.1.1"],
  //         },
  //       ],
  //       tags: ["soc2-type1", "tenancy"],
  //       targets: {
  //         kinds: [
  //           "Deployment",
  //           "StatefulSet",
  //           "ReplicaSet",
  //           "ReplicationController",
  //           "HorizontalPodAutoscaler",
  //         ],
  //         labels: [],
  //         namespaces: [],
  //       },
  //       tenant: "",
  //     },
  //   ],
  //   total: 1,
  //   errors: [],
  // };

  console.log(data);
  return (
    <Page
      error={error || data?.errors}
      loading={isLoading}
      className={className}
    >
      {data?.policies && <PolicyTable policies={data.policies} />}
    </Page>
  );
}

export default styled(PoliciesList).attrs({
  className: PoliciesList.name,
})``;
