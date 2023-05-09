import * as React from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { darcula } from "react-syntax-highlighter/dist/cjs/styles/prism";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { ObjectRef } from "../lib/api/core/types.pb";
import { createYamlCommand } from "../lib/utils";
import CopyToClipboard from "./CopyToCliboard";

export type YamlViewProps = {
  className?: string;
  yaml: string;
  object?: ObjectRef;
};

const YamlHeader = styled.div`
  background: ${(props) => props.theme.colors.neutralGray};
  padding: ${(props) => props.theme.spacing.small};
  width: 100%;
  border-bottom: 1px solid ${(props) => props.theme.colors.neutral20};
  font-family: monospace;
  color: ${(props) => props.theme.colors.primary10};
  text-overflow: ellipsis;
`;

function UnstyledYamlView({ yaml, object, className }: YamlViewProps) {
  const headerText = createYamlCommand(
    object.kind,
    object.name,
    object.namespace
  );

  const { settings } = React.useContext(AppContext);
  const dark = settings.theme === "dark";

  const styleProps = {
    customStyle: {
      margin: 0,
      ...(!dark && { backgroundColor: "transparent" }),
    },

    codeTagProps: {
      style: {
        wordBreak: "break-word",
      },
    },

    lineProps: { style: { flexWrap: "wrap" } },

    ...(dark && { style: darcula }),
  };

  return (
    <div className={className}>
      <YamlHeader>
        {headerText}
        {headerText && (
          <CopyToClipboard
            value={headerText}
            className="yaml-copy"
            size="small"
          />
        )}
      </YamlHeader>
      <SyntaxHighlighter
        language="yaml"
        {...styleProps}
        wrapLongLines
        wrapLines
        showLineNumbers
      >
        {yaml}
      </SyntaxHighlighter>
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
    padding: ${(props) => props.theme.spacing.medium}
      ${(props) => props.theme.spacing.small} !important;
  }
`;

export const DialogYamlView = styled(YamlView)`
  margin-bottom: 0;
  overflow: auto;
  overflow-x: hidden;
`;

export default YamlView;
