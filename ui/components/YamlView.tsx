import * as React from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { darcula } from "react-syntax-highlighter/dist/cjs/styles/prism";
import styled from "styled-components";
import { ThemeTypes } from "../contexts/AppContext";
import { useInDarkMode } from "../hooks/theme";
import CopyToClipboard from "./CopyToCliboard";
import Flex from "./Flex";

export type YamlViewProps = {
  className?: string;
  type?: string;
  yaml: string;
  header?: string;
  theme?: ThemeTypes;
};

const YamlHeader = styled(Flex)`
  background: ${(props) => props.theme.colors.neutralGray};
  padding: ${(props) => props.theme.spacing.small};
  border-bottom: 1px solid ${(props) => props.theme.colors.neutral20};
  font-family: monospace;
  color: ${(props) => props.theme.colors.primary10};
  text-overflow: ellipsis;
`;

function UnstyledYamlView({
  yaml,
  header,
  className,
  theme,
  type = "yaml",
}: YamlViewProps) {
  const dark = theme ? theme === ThemeTypes.Dark : useInDarkMode();

  const styleProps = {
    customStyle: {
      margin: 0,
      backgroundColor: "transparent",
    },

    codeTagProps: {
      style: {
        wordBreak: "break-word",
      },
    },

    lineProps: { style: { textWrap: "wrap" } },

    ...(dark && { style: darcula }),
  };

  return (
    <div className={className}>
      {header && (
        <YamlHeader wide gap="4" alignItems="center">
          {header}
          <CopyToClipboard
            value={header}
            className="yamlheader-copy"
            size="small"
          />
        </YamlHeader>
      )}

      <div className="code-wrapper">
        <div className="copy-wrapper">
          <CopyToClipboard value={yaml} className="yaml-copy" size="base" />
        </div>
        <SyntaxHighlighter
          language={type}
          {...styleProps}
          wrapLines
          showLineNumbers
        >
          {yaml}
        </SyntaxHighlighter>
      </div>
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

  .code-wrapper {
    position: relative;
  }
  .copy-wrapper {
    position: absolute;
    right: 4px;
    top: 8px;
    background: ${(props) => props.theme.colors.neutralGray};
    padding: 4px 8px;
    border-radius: 2px;
  }
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
