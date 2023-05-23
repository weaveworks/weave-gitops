import React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

import styled from "styled-components";
import { useGetPolicyValidationDetails } from "../../hooks/policyViolations";
import { PolicyValidation } from "../../lib/api/core/core.pb";
import { Kind } from "../../lib/api/core/types.pb";
import { formatURL } from "../../lib/nav";
import { V2Routes } from "../../lib/types";
import Flex from "../Flex";
import Link from "../Link";
import Page from "../Page";
import Text from "../Text";
import Timestamp from "../Timestamp";

import YamlView from "../YamlView";
import Parameters from "./Parameters";
import Severity from "./Severity";

const SectionWrapper = ({ tilte, children }) => {
  return (
    <Flex column wide gap="8" data-testid="occurrences">
      <Text bold color="neutral30">
        {tilte}
      </Text>
      {children}
    </Flex>
  );
};

interface IViolationDetailsProps {
  violation: PolicyValidation;
}
const ViolationDetails = ({ violation }: IViolationDetailsProps) => {
  const {
    severity,
    createdAt,
    category,
    howToSolve,
    description,
    violatingEntity,
    entity,
    namespace,
    occurrences,
    name,
    parameters,
  } = violation || {};

  const violatingEntityObj = JSON.parse(violatingEntity);
  const entityKind = violatingEntityObj.kind;

  const headers = [
    {
      rowkey: "Policy Name",
      value: name,
    },
    {
      rowkey: "Violation Time",
      value: <Timestamp time={createdAt} />,
    },
    {
      rowkey: "Severity",
      children: <Severity severity={severity || ""} />,
    },
    {
      rowkey: "Category",
      value: category,
    },
    {
      rowkey: "Application",
      children: (
        <Link
          to={formatURL(
            entityKind === Kind.Kustomization
              ? V2Routes.Kustomization
              : V2Routes.HelmRelease,
            {
              name: entity,
              namespace: namespace,
            }
          )}
        >
          {namespace}/{entity}
        </Link>
      ),
    },
  ];

  return (
    <Flex wide tall column gap="32">
      <Flex column gap="8">
        {headers.map((h) => {
          return (
            <Flex center gap="8" data-testid={h.rowkey} key={h.rowkey}>
              <Text color="neutral30" semiBold size="medium">
                {h.rowkey}:
              </Text>
              <Text color="neutral40" size="medium">
                {h.children || h.value || "-"}
              </Text>
            </Flex>
          );
        })}
      </Flex>
      <SectionWrapper tilte={` Occurrences ( ${occurrences?.length} )`}>
        <ul className="occurrences">
          {occurrences?.map((item) => (
            <li key={item.message}>
              <Text size="medium"> {item.message}</Text>
            </li>
          ))}
        </ul>
      </SectionWrapper>
      <SectionWrapper tilte="Description:">
        <ReactMarkdown children={description || ""} className="editor" />
      </SectionWrapper>
      <SectionWrapper tilte="How to solve:">
        <ReactMarkdown
          children={howToSolve || ""}
          remarkPlugins={[remarkGfm]}
          className="editor"
        />
      </SectionWrapper>
      <SectionWrapper tilte="Violating Entity:">
        <YamlView
          type="json"
          yaml={violatingEntity}
          object={null}
          className="code"
        />
      </SectionWrapper>
      <SectionWrapper tilte=" Parameters Values:">
        <Parameters parameters={parameters} />
      </SectionWrapper>
    </Flex>
  );
};

interface Props {
  id: string;
  clusterName?: string;
  className?: string;
}

const PolicyViolationDetails = ({ id, className }: Props) => {
  const { data, error, isLoading } = useGetPolicyValidationDetails({
    violationId: id,
  });
  return (
    <Page error={error} loading={isLoading} className={className}>
      {data && <ViolationDetails violation={data.violation} />}
    </Page>
  );
};

export default styled(PolicyViolationDetails)`
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
  }
`;
