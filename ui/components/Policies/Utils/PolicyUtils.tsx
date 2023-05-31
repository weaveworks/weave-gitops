import { Chip } from "@material-ui/core";
import * as React from "react";
import ReactMarkdown from "react-markdown/react-markdown.min";
import styled from "styled-components";
import Flex from "../../Flex";
import Text from "../../Text";

export const parseValue = (parameter: {
  type?: string | undefined;
  value?: any;
}) => {
  if (!parameter.value) return <ChipWrap label="undefined" />;
  switch (parameter.type) {
    case "boolean":
      return parameter.value.value ? "true" : "false";
    case "array":
      return parameter.value.value.join(", ");
    case "string":
      return parameter.value.value;
    case "integer":
      return parameter.value.value.toString();
  }
};

export const SectionWrapper = ({ title, children }) => {
  return (
    <Flex column wide gap="8" data-testid="occurrences">
      <Text bold color="neutral30">
        {title}
      </Text>
      {children}
    </Flex>
  );
};

export const ChipWrap = styled(Chip)`
  &.MuiChip-root {
    color: ${(props) => props.theme.colors.black};
    background-color: ${(props) => props.theme.colors.neutralGray};
    padding: 2px 4px;
    height: inherit;
    border-radius: 4px;
  },
  .MuiChip-label {
    padding: 0;
  }
`;

export const Editor = styled(ReactMarkdown)`
  width: calc(100% - 24px);
  padding: ${(props) => props.theme.spacing.small};
  overflow: scroll;
  background: ${(props) => props.theme.colors.neutralGray};
  max-height: 300px;
  & a {
    color: ${(props) => props.theme.colors.primary};
  },
  & > *:first-child {
    margin-top: ${(props) => props.theme.spacing.none};
  },
  & > *:last-child {
    margin-bottom: ${(props) => props.theme.spacing.none};
  }
`;
