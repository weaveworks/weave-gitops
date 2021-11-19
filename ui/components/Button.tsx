// eslint-disable-next-line
import { CircularProgress } from "@material-ui/core";
import MaterialButton, { ButtonProps } from "@material-ui/core/Button";
import * as React from "react";
import styled from "styled-components";

type Props = ButtonProps & {
  loading?: boolean;
  endIcon?: React.ReactNode;
  className?: string;
};

function Button(props: Props) {
  return (
    <MaterialButton
      variant="outlined"
      color="primary"
      disabled={props.loading}
      endIcon={props.loading ? <CircularProgress size={16} /> : props.endIcon}
      {...props}
    />
  );
}

export default styled(Button)`
  display: flex;
  justify-content: space-evenly;
  &.auth-button {
    &.MuiButton-contained {
      background-color: black;
      color: white;
    }
  }
`;
