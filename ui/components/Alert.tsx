import { Alert as MaterialAlert, AlertTitle } from "@material-ui/lab";
import * as React from "react";
import styled from "styled-components";
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
        icon={
          <Icon type={IconType.ErrorIcon} size="medium" color="alertDark" />
        }
        severity={severity}
      >
        <AlertTitle>{title}</AlertTitle>
        <Text color="black">{message}</Text>
      </MaterialAlert>
    </Flex>
  );
}

const Alert = styled(UnstyledAlert)`
  .MuiAlert-standardError {
    background-color: ${(props) => props.theme.colors.alertLight};
  }
  .MuiAlertTitle-root {
    color: ${(props) => props.theme.colors.black};
  }
`;

export default Alert;
