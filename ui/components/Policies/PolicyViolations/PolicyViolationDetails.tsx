import React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

import styled from "styled-components";
import { useGetPolicyValidationDetails } from "../../../hooks/policyViolations";
import { PolicyValidation } from "../../../lib/api/core/core.pb";
import { Kind } from "../../../lib/api/core/types.pb";
import { formatURL } from "../../../lib/nav";
import { V2Routes } from "../../../lib/types";
import Flex from "../../Flex";
import Link from "../../Link";
import Page from "../../Page";
import Text from "../../Text";
import Timestamp from "../../Timestamp";

import { AppContext } from "../../../contexts/AppContext";
import { FluxObject } from "../../../lib/objects";
import HeaderRows, { Header } from "../Utils/HeaderRows";
import Parameters from "../Utils/Parameters";
import Severity from "../Utils/Severity";

const SectionWrapper = ({ title, children }) => {
  return (
    <Flex column wide gap="8" data-testid="occurrences">
      <Text bold color="neutral30">
        {title}
      </Text>
      {children}
    </Flex>
  );
};

interface IViolationDetailsProps {
  violation: PolicyValidation;
  kind: string;
}
const ViolationDetails = ({ violation, kind }: IViolationDetailsProps) => {
  const { setDetailModal } = React.useContext(AppContext);
  const {
    clusterName,
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

  const entityObject = new FluxObject({ payload: violatingEntity });
  const headers: Header[] = [
    {
      rowkey: "Policy Name",
      value: name,
    },
    {
      rowkey: "Application",
      children: (
        <Link
          to={formatURL(
            Kind[kind] === Kind.Kustomization
              ? V2Routes.Kustomization
              : V2Routes.HelmRelease,
            {
              name: entity,
              namespace: namespace,
              clusterName,
            }
          )}
        >
          {namespace}/{entity}
        </Link>
      ),
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
      rowkey: "Violating Entity",
      children: (
        <Text
          pointer
          size="medium"
          color="primary"
          onClick={() => setDetailModal({ object: entityObject })}
        >
          {entityObject.namespace}/{entityObject.name}
        </Text>
      ),
    },
  ];

  return (
    <Flex wide tall column gap="32">
      <HeaderRows headers={headers} />
      <SectionWrapper title={` Occurrences ( ${occurrences?.length} )`}>
        <ul className="occurrences">
          {occurrences?.map((item) => (
            <li key={item.message}>
              <Text size="medium"> {item.message}</Text>
            </li>
          ))}
        </ul>
      </SectionWrapper>
      <SectionWrapper title="Description:">
        <ReactMarkdown children={description || ""} className="editor" />
      </SectionWrapper>
      <SectionWrapper title="How to solve:">
        <ReactMarkdown
          children={howToSolve || ""}
          remarkPlugins={[remarkGfm]}
          className="editor"
        />
      </SectionWrapper>
      <SectionWrapper title=" Parameters Values:">
        <Parameters parameters={parameters} />
      </SectionWrapper>
    </Flex>
  );
};

interface Props {
  id: string;
  clusterName?: string;
  className?: string;
  kind?: string;
}

const PolicyViolationDetails = ({ id, className, kind }: Props) => {
  const { data, error, isLoading } = useGetPolicyValidationDetails({
    violationId: id,
  });
  return (
    <Page error={error} loading={isLoading} className={className}>
      {data && <ViolationDetails violation={data.violation} kind={kind} />}
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
    margin: 0;
  }
`;
