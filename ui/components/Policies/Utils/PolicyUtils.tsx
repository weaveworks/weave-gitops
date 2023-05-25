import { Chip } from "@material-ui/core";
import React from "react";
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
    background-color: ${(props) => props.theme.colors.neutralGray};
    padding: 2px 4px;
    height: inherit;
    border-radius: 4px;
  }
  .MuiChip-label {
    padding: 0;
  }
`;
