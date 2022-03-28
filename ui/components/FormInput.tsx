import { FormControl, FormLabel } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { FormContext } from "./ControlledForm";
import Input, { InputProps } from "./Input";
import Text from "./Text";

export type FormInputProps = {
  name: string;
  className?: string;
  label: string;
  component?: any;
  valuePropName?: string;
} & InputProps;

export const Label = styled(({ className, children, name, required }) => (
  <FormLabel className={className} htmlFor={name} required={required}>
    <Text size="small" bold>
      {children}
    </Text>
  </FormLabel>
))`
  ${Text} {
    text-transform: uppercase;
  }

  label {
    margin-bottom: 4px;
  }
`;

function FormInput({
  className,
  label,
  name,
  component,
  valuePropName = "value",
  ...props
}: FormInputProps) {
  const Component = component || Input;

  return (
    <FormContext.Consumer>
      {({ handleChange, findValue }) => (
        <FormControl className={className}>
          <Label name={name} required={props.required}>
            {label}
          </Label>
          <Component
            id={name}
            variant="outlined"
            {...props}
            {...{ [valuePropName]: findValue(name) }}
            onChange={(ev) => {
              const v = ev.target[valuePropName];
              handleChange(name, v);
            }}
          />
        </FormControl>
      )}
    </FormContext.Consumer>
  );
}

export default styled(FormInput).attrs({ className: FormInput.name })`
  min-width: 300px;
`;
