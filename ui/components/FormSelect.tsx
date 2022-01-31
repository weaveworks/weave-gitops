import { FormHelperText, Select } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import FormInput, { FormInputProps } from "./FormInput";

type Props = {
  className?: string;
} & FormInputProps;

function FormSelect({ helperText, ...props }: Props) {
  return (
    <>
      <FormInput {...props} component={Select} />
      <FormHelperText>{helperText}</FormHelperText>
    </>
  );
}

export default styled(FormSelect).attrs({ className: FormSelect.name })`
  .MuiSelect-root {
    min-width: 300px;
  }
`;
