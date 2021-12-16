import {
  Alert as MaterialAlert,
  // eslint-disable-next-line
  AlertProps,
  AlertTitle,
} from "@material-ui/lab";
import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";

/** Alert Properties */
export interface Props {
  /** string of one of the colors from our `MuiTheme` - also sets the corresponding material icon - see /ui/lib/theme.ts and https://mui.com/customization/theming/ */
  severity?: AlertProps["severity"];
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
      <MaterialAlert severity={severity}>
        <AlertTitle>{title}</AlertTitle>
        {message}
      </MaterialAlert>
    </Flex>
  );
}

const Alert = styled(UnstyledAlert)``;

export default Alert;
