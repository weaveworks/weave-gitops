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
    font-weight: 600;
  }
  h4 {
    margin: 0;
  }
  &.table {
    &.MuiButton-text {
      min-width: 0px;
    }
  }
  &.lowercase {
    text-transform: none;
  }
`;
