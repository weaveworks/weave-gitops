// eslint-disable-next-line
import MaterialButton, { ButtonProps } from "@material-ui/core/Button";
import * as React from "react";
import styled from "styled-components";

type Props = ButtonProps & {
  className?: string;
};

function Button(props: Props) {
  return <MaterialButton {...props} />;
}

export default styled(Button)`
  &.applications {
    border-radius: 0;
    border-color: ${(props) => props.theme.colors.black};
  }
  p {
    margin: 0px;
  }
  &.table {
    padding: 0px 4px 0px 0px;
    &.MuiButton-text {
      min-width: 0px;
    }
  }
  &.bold {
    font-weight: 600;
  }
  &.lowercase {
    text-transform: none;
  }
`;
