import { Checkbox } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import FormInput, { FormInputProps } from "./FormInput";

type Props = FormInputProps;

function FormCheckbox(props: Props) {
  return (
    <FormInput
      {...props}
      //   Covert to bool to satisfy mui Checkbox props
      component={(p) => (
        <Checkbox {...p} checked={p.checked === "true" || Boolean(p.checked)} />
      )}
      valuePropName="checked"
    />
  );
}

export default styled(FormCheckbox).attrs({ className: FormCheckbox.name })``;
