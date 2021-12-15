// eslint-disable-next-line
import { CircularProgress } from "@material-ui/core";
import MaterialButton, { ButtonProps } from "@material-ui/core/Button/Button";
import * as React from "react";
import styled from "styled-components";
import { theme } from "..";

/** Button Properties */
export interface Props extends ButtonProps {
  /** Changes the Buttons `startIcon` prop to Mui's `<CircularProgress />` and sets `disabled` to `true`. */
  loading?: boolean;
  /** `<Icon />` Element to come after `<Button />` content. */
  startIcon?: React.ReactNode;
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
      startIcon={
        props.loading ? <CircularProgress size={16} /> : props.startIcon
      }
      disableElevation={true}
      {...props}
    />
  );
}

const Button = styled(UnstyledButton)`
  &.MuiButton-root {
    line-height: 1;
    border-radius: 2px;
    font-weight: 600;
  }
  &.MuiButton-outlined {
    padding: 8px 12px;
    border-color: ${theme.colors.neutral20};
  }
`;

export default Button;
