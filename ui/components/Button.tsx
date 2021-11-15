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

export default styled(Button)``;
