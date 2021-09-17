import {
  Alert as MaterialAlert,
  // eslint-disable-next-line
  AlertProps,
  AlertTitle,
} from "@material-ui/lab";
import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";

type Props = {
  className?: string;
  title?: string;
  message?: string | JSX.Element;
  severity?: AlertProps["severity"];
};

function Alert({ className, title, message, severity }: Props) {
  return (
    <Flex wide className={className}>
      <MaterialAlert severity={severity}>
        <AlertTitle>{title}</AlertTitle>
        {message}
      </MaterialAlert>
    </Flex>
  );
}

export default styled(Alert)``;
