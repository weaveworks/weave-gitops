import {
  TextField,
  // eslint-disable-next-line
  TextFieldProps,
} from "@mui/material";
import * as React from "react";
import styled from "styled-components";

export type InputProps = TextFieldProps;

function Input({ ...props }: InputProps) {
  return <TextField {...props} />;
}

export default styled(Input).attrs({ className: Input.name })``;
