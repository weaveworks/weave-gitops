// eslint-disable-next-line
import { CircularProgress } from "@material-ui/core";
import MaterialButton, { ButtonProps } from "@material-ui/core/Button/Button";
import * as React from "react";
import styled from "styled-components";

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
function UnstyledButton({ loading, ...props }: Props) {
  return (
    <MaterialButton
      variant={props.variant || "outlined"}
      color={props.color || "primary"}
      disabled={loading}
      startIcon={loading ? <CircularProgress size={16} /> : props.startIcon}
      disableElevation={true}
      {...props}
    />
  );
}

const Button = styled(UnstyledButton)`
  &.MuiButton-root {
    line-height: 1;
    border-radius: ${(props) => props.theme.borderRadius.soft};
    font-weight: 600;
  }
  &.MuiButton-outlined {
    padding: 8px 12px;
    border-color: ${(props) => props.theme.colors.neutral20};
  }
`;

export default Button;
