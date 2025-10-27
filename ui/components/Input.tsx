import { TextField } from "@mui/material";
import * as React from "react";
import styled from "styled-components";

export type InputProps = React.ComponentProps<typeof TextField>;

function Input({ ...props }: InputProps) {
  return <TextField {...props} />;
}

export default styled(Input).attrs({ className: Input.name })``;
