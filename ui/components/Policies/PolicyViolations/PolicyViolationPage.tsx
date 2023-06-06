import React from "react";

import styled from "styled-components";
import { useGetPolicyValidationDetails } from "../../../hooks/policyViolations";
import { Kind } from "../../../lib/api/core/types.pb";
import { formatURL } from "../../../lib/nav";
import { V2Routes } from "../../../lib/types";
import Page from "../../Page";

import { PolicyValidation } from "../../../lib/api/core/core.pb";
import { FluxObject } from "../../../lib/objects";
import { ViolationDetails } from "./PolicyViolationDetails";

const getPath = (kind: string, violation?: PolicyValidation) => {
  if (!violation) return [];
  const { name, entity, namespace, clusterName, policyId } = violation;
  if (kind === Kind.Policy) {
    const policyUrl = formatURL(`${V2Routes.PolicyDetailsPage}/violations`, {
      id: policyId,
      clusterName,
      name,
    });
    return [
      { label: "Policies", url: V2Routes.Policies },
      { label: name, url: policyUrl },
    ];
  }
  const entityUrl = formatURL(
    Kind[kind] === Kind.Kustomization
      ? `${V2Routes.Kustomization}/violations`
      : `${V2Routes.HelmRelease}/violations`,
    {
      name: entity,
      namespace: namespace,
      clusterName: clusterName,
    }
  );
  return [
    { label: "Applications", url: V2Routes.Automations },
    { label: entity, url: entityUrl },
  ];
};

interface Props {
  id: string;
  name: string;
  clusterName?: string;
  className?: string;
  kind?: string;
}

const PolicyViolationPage = ({
  id,
  name,
  className,
  clusterName,
  kind,
}: Props) => {
  const { data, error, isLoading } = useGetPolicyValidationDetails({
    validationId: id,
    clusterName,
  });

  const violation = data?.validation;
  const entityObject = new FluxObject({
    payload: violation?.violatingEntity,
  });

  return (
    <Page
      error={error}
      loading={isLoading}
      className={className}
      path={[...getPath(kind, data?.validation), { label: name || "" }]}
    >
      {data && (
        <ViolationDetails
          violation={data.validation}
          entityObject={entityObject}
          kind={kind}
        />
      )}
    </Page>
  );
};

export default styled(PolicyViolationPage)`
  ul.occurrences {
    padding-left: ${(props) => props.theme.spacing.base};
    margin: 0;
  }
`;
