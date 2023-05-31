import React from "react";
import { Chip } from "@material-ui/core";
import styled from "styled-components";

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
