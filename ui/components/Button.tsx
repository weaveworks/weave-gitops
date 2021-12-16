// eslint-disable-next-line
import { CircularProgress } from "@material-ui/core";
import MaterialButton, { ButtonProps } from "@material-ui/core/Button/Button";
import * as React from "react";
import styled from "styled-components";
import { muiTheme } from "../lib/theme";

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
function UnstyledButton(props: Props) {
  return (
    <MaterialButton
      variant="outlined"
      color="primary"
      disabled={props.loading}
      endIcon={props.loading ? <CircularProgress size={16} /> : props.endIcon}
      disableElevation={true}
      {...props}
    />
  );
}
const Button = styled(UnstyledButton)`
  &.MuiButton-outlined {
    border-color: ${muiTheme.palette.text.secondary};
  }
`;

export default Button;
