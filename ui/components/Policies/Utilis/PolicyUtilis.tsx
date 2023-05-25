import * as React from "react";
import styled from "styled-components";
import Flex from "../../Flex";
import Text from "../../Text";
import { Chip } from "@material-ui/core";
import ReactMarkdown from "react-markdown";

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
    default:
  }
};

export const ParameterCell = ({
  label,
  value,
}: {
  label: string;
  value: string | undefined;
}) => {
  return (
    <Flex wide column data-testid={label} gap="4">
      <Text color="neutral30">{label}</Text>
      <Text color="black">{value}</Text>
    </Flex>
  );
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

export const ModeWrapper = styled(Flex)`
  align-items: center;
  justify-content: flex-start;
  display: inline-flex;
  margin-right: ${(props) => props.theme.spacing.xs};
  svg {
    color: ${(props) => props.theme.colors.neutral30};
    font-size: ${(props) => props.theme.fontSizes.large};
    margin-right: ${(props) => props.theme.spacing.xxs};
  }
  span {
    text-transform: capitalize;
  }
`;

export const ParameterWrapper = styled(Flex)`
  border: 1px solid ${(props) => props.theme.colors.neutral20};
  box-sizing: border-box;
  border-radius: ${(props) => props.theme.spacing.xxs};
  padding: ${(props) => props.theme.spacing.base};
  width: 100%;
`;

export const ChipWrap = styled(Chip)`
  &.MuiChip-root {
    color: ${(props) => props.theme.colors.black};
    background-color: ${(props) => props.theme.colors.neutral10};
    padding: 2px 4px;
    height: inherit;
    border-radius: 4px;
  }
  .MuiChip-label {
    padding: 0;
  }
`;

export const Editor = styled(ReactMarkdown)`
  width: calc(100% - 24px);
  padding: ${(props) => props.theme.spacing.small};
  overflow: scroll;
  background: ${(props) => props.theme.colors.neutral10};
  max-height: 300px;
  & a {
    color: ${(props) => props.theme.colors.primary};
  }
  ,
  & > *:first-child {
    margin-top: ${(props) => props.theme.spacing.none};
  }
  ,
  & > *:last-child {
    margin-bottom: ${(props) => props.theme.spacing.none};
  }
`;
