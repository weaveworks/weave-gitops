// eslint-disable-next-line
import { CircularProgress } from "@material-ui/core";
import MaterialButton, { ButtonProps } from "@material-ui/core/Button/Button";
import * as React from "react";
import styled from "styled-components";

/** Button Properties */
export interface Props extends ButtonProps {
  /** Changes the Buttons `endIcon` prop to Mui's `<CircularProgress />` and sets `disabled` to `true`. */
  loading?: boolean;
  /** `<Icon />` Element to come after `<Button />` content. */
  endIcon?: React.ReactNode;
  /** CSS MUI Overrides or other styling. */
  className?: string;
}

/** Form Button */
function Button({ loading, ...props }: Props) {
  return (
    <MaterialButton
      variant="outlined"
      color="primary"
      disabled={loading}
      endIcon={loading ? <CircularProgress size={16} /> : props.endIcon}
      disableElevation={true}
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
  &.borderless {
    &.MuiButton-outlined {
      border: none;
    }
  }
`;
