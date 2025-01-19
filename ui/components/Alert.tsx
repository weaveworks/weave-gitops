import { AlertTitle, Alert as MaterialAlert } from "@mui/material";
import * as React from "react";
import type { JSX } from "react";
import styled from "styled-components";
import { ThemeTypes } from "../contexts/AppContext";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Text from "./Text";

/** Alert Properties */
export interface Props {
  /** string of one of the colors from our `MuiTheme` - also sets the corresponding material icon - see /ui/lib/theme.ts and https://mui.com/customization/theming/ */
  severity?: "error" | "info" | "success" | "warning";
  /** Overrides `justify-content: flex-start` (left) to render the Alert in the center of it's 100% width `<Flex />` component */
  center?: boolean;
  /** text for Mui's `<AlertTitle />` component */
  title?: string;
  /** Appears under `title` */
  message?: string | JSX.Element;
  /** CSS MUI Overrides or other styling */
  className?: string;
}
/** Form Alert */
function UnstyledAlert({ center, title, message, severity, className }: Props) {
  return (
    <Flex wide start={!center} className={className}>
      <MaterialAlert
        icon={<Icon type={IconType.ErrorIcon} size="medium" />}
        severity={severity}
      >
        <AlertTitle>{title}</AlertTitle>
        <Text color="neutral40">{message}</Text>
      </MaterialAlert>
    </Flex>
  );
}

const Alert = styled(UnstyledAlert)`
  .MuiAlert-root {
    border-radius: 8px;
    line-height: 16px;
  }
  .MuiAlert-standardError {
    svg {
      color: ${(props) => props.theme.colors.alertDark};
    }
    background-color: ${(props) => props.theme.colors.alertLight};
  }
  .MuiAlertTitle-root {
    color: ${(props) => props.theme.colors.neutral40};
  }
  .MuiAlert-standardInfo {
    svg {
      color: ${(props) => props.theme.colors.primaryLight10};
    }
    background-color: ${(props) =>
      props.theme.mode === ThemeTypes.Dark
        ? props.theme.colors.primary20
        : props.theme.colors.primaryLight05};
  }
`;

export default Alert;
