import * as React from "react";
import styled from "styled-components";
import { ObjectRef } from "../lib/api/core/types.pb";
import { fluxObjectKindToKind } from "../lib/objects";
import { IconButton } from "./Button";
import Icon, { IconType } from "./Icon";
type Props = {
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

const CopyButton = styled(IconButton)`
  &.MuiButton-outlinedPrimary {
    border: 1px solid ${(props) => props.theme.colors.neutral10};
    padding: ${(props) => props.theme.spacing.xs};
  }
  &.MuiButton-root {
    height: initial;
    width: initial;
    min-width: 0px;
  }
`;

function YamlView({ yaml, object, className }: Props) {
  const [copied, setCopied] = React.useState(false);
  const headerText = `kubectl get ${fluxObjectKindToKind(
    object.kind
  ).toLowerCase()} ${object.name} -n ${object.namespace} -o yaml `;

  return (
    <div className={className}>
      <YamlHeader>
        {headerText}
        <CopyButton
          onClick={() => {
            navigator.clipboard.writeText(headerText);
            setCopied(true);
          }}
        >
          <Icon type={IconType.FileCopyIcon} size="small" />
        </CopyButton>
        {copied && " Copied!"}
      </YamlHeader>
      <pre>
        {yaml.split("\n").map((yaml, index) => (
          <code key={index}>{yaml}</code>
        ))}
      </pre>
    </div>
  );
}

export default styled(YamlView).attrs({
  className: YamlView.name,
})`
  width: calc(100% - ${(props) => props.theme.spacing.medium});
  font-size: ${(props) => props.theme.fontSizes.small};
  border: 1px solid ${(props) => props.theme.colors.neutral20};
  border-radius: 8px;
  overflow: scroll;
  pre {
    padding: ${(props) => props.theme.spacing.small};
    white-space: pre-wrap;
  }

  pre::before {
    counter-reset: listing;
  }

  code {
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
