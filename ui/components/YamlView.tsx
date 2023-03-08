import * as React from "react";
import styled from "styled-components";
import { ObjectRef } from "../lib/api/core/types.pb";
import CopyToClipboard from "./CopyToCliboard";

export type YamlViewProps = {
  className?: string;
  yaml: string;
  object?: ObjectRef;
};
const YamlHeader = styled.div`
  background: ${(props) => props.theme.colors.neutral10};
  padding: ${(props) => props.theme.spacing.small};
  width: 100%;
  border-bottom: 1px solid ${(props) => props.theme.colors.neutral20};
  font-family: monospace;
  color: ${(props) => props.theme.colors.primary10};
  text-overflow: ellipsis;
`;

function UnstyledYamlView({ yaml, object, className }: YamlViewProps) {
  const headerText = `kubectl get ${object.kind.toLowerCase()} ${
    object.name
  } -n ${object.namespace} -o yaml `;

  return (
    <div className={className}>
      <YamlHeader>
        {headerText}
        <CopyToClipboard
          value={headerText}
          className="yaml-copy"
          size="small"
        ></CopyToClipboard>
      </YamlHeader>
      <pre>
        {yaml.split("\n").map((yaml, index) => (
          <code key={index}>{yaml}</code>
        ))}
      </pre>
    </div>
  );
}

const YamlView = styled(UnstyledYamlView).attrs({
  className: UnstyledYamlView.name,
})`
  margin-bottom: ${(props) => props.theme.spacing.small};
  width: 100%;
  font-size: ${(props) => props.theme.fontSizes.small};
  border: 1px solid ${(props) => props.theme.colors.neutral20};
  border-radius: 8px;
  overflow: hidden;
  pre {
    overflow-y: scroll;
    overflow-x: hidden;
    padding: ${(props) => props.theme.spacing.small};
    white-space: pre-wrap;
  }

  pre::before {
    counter-reset: listing;
  }

  code {
    width: 100%;
    counter-increment: listing;
    text-align: left;
    float: left;
    clear: left;
  }

  code::before {
    width: 28px;
    color: ${(props) => props.theme.colors.primary10};
    content: counter(listing);
    display: inline-block;
    float: left;
    height: auto;
    padding-left: auto;
    margin-right: ${(props) => props.theme.spacing.small};
    text-align: right;
  }
`;

export const DialogYamlView = styled(YamlView)`
  margin-bottom: 0;
  overflow: auto;
  overflow-x: hidden;
`;

export default YamlView;
