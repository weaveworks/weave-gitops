import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  yaml: string;
};

function YamlView({ yaml, className }: Props) {
  console.log(yaml.split("\n"));

  return (
    <pre className={className}>
      {yaml.split("\n").map((yaml, index) => (
        <code key={index}>{yaml}</code>
      ))}
    </pre>
  );
}

export default styled(YamlView).attrs({
  className: YamlView.name,
})`
  width: calc(100% - ${(props) => props.theme.spacing.medium});
  font-size: ${(props) => props.theme.fontSizes.small};
  border: 1px solid ${(props) => props.theme.colors.neutral20};
  border-radius: 8px;
  padding: ${(props) => props.theme.spacing.small};
  overflow: scroll;
  pre {
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
