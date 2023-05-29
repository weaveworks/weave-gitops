import React from "react";

import styled from "styled-components";
import { useGetPolicyValidationDetails } from "../../../hooks/policyViolations";
import { Kind } from "../../../lib/api/core/types.pb";
import { formatURL } from "../../../lib/nav";
import { V2Routes } from "../../../lib/types";
import Page from "../../Page";

import { FluxObject } from "../../../lib/objects";
import { ViolationDetails } from "./PolicyViolationDetails";

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

  const entityUrl = formatURL(
    Kind[kind] === Kind.Kustomization
      ? `${V2Routes.Kustomization}/violations`
      : `${V2Routes.HelmRelease}/violations`,
    {
      name: violation?.entity,
      namespace: violation?.namespace,
      clusterName: violation?.clusterName,
    }
  );
  return (
    <Page
      error={error}
      loading={isLoading}
      className={className}
      path={[
        { label: "Applications", url: V2Routes.Automations },
        { label: violation?.entity, url: entityUrl },
        { label: name || "" },
      ]}
    >
      {data && (
        <ViolationDetails
          violation={data.validation}
          entityUrl={entityUrl}
          entityObject={entityObject}
        />
      )}
    </Page>
  );
};

export default styled(PolicyViolationPage)`
  .editor {
    & a {
      color: ${(props) => props.theme.colors.primary};
    }
    ,
    & > *:first-child {
      margin-top: 0;
    }
    ,
    & > *:last-child {
      margin-bottom: 0;
    }

    width: calc(100% - 24px);
    padding: 12px;
    overflow: scroll;
    background: ${(props) => props.theme.colors.neutralGray};
    max-height: 300px;
  }
  .code {
    pre {
      max-height: 300px;
      overflow: auto;
    }
    code > span {
      flex-wrap: wrap;
    }
  }
  ul.occurrences {
    padding-left: ${(props) => props.theme.spacing.base};
    margin: 0;
  }
`;
