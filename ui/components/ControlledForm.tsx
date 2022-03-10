import _ from "lodash";
import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  children?: any;
  onSubmit?: (state: any) => any;
  onChange?: (name: string, value: string) => any;
  state: { values: any };
};

interface FormContextType {
  handleChange: (name: string, value: string) => void;
  findValue: (name: string) => string;
}

export const FormContext = React.createContext<FormContextType>(null);

function ControlledForm({
  className,
  children,
  onSubmit,
  onChange,
  state,
}: Props) {
  return (
    <FormContext.Provider
      value={{
        findValue: (name: string) => _.get(state.values, name),
        handleChange: (name: string, value: string) => onChange(name, value),
      }}
    >
      <form
        onSubmit={(ev) => {
          ev.preventDefault();
          onSubmit(state.values);
        }}
        className={className}
      >
        {children}
      </form>
    </FormContext.Provider>
  );
}

export default styled(ControlledForm).attrs({
  className: ControlledForm.name,
})``;
