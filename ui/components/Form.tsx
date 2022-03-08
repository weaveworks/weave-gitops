import _ from "lodash";
import * as React from "react";
import styled from "styled-components";

interface Props {
  className?: string;
  children?: any;
  onSubmit?: (state: any) => any;
  onChange?: (state: any) => any;
  initialState: any;
}

interface FormContextType {
  handleChange: (name: string, value) => void;
  findValue: (name: string) => string;
}

export const FormContext = React.createContext<FormContextType>(null);

enum ActionTypes {
  setValue,
}
type Action = { type: ActionTypes.setValue; value: string; name: string };

function reducer(state: any, action: Action) {
  switch (action.type) {
    case ActionTypes.setValue:
      return {
        ...state,
        values: {
          ...state.values,
          [action.name]: action.value,
        },
      };

    default:
      break;
  }
  return state;
}

function Form({
  className,
  children,
  onSubmit,
  initialState,
  onChange,
}: Props) {
  const [state, dispatch] = React.useReducer(reducer, { values: initialState });

  React.useEffect(() => {
    if (onChange) {
      onChange(state);
    }
  }, [state]);

  return (
    <FormContext.Provider
      value={{
        findValue: (name: string) => _.get(state.values, name),
        handleChange: (name: string, value: string) =>
          dispatch({ type: ActionTypes.setValue, name, value }),
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

export default styled(Form).attrs({ className: Form.name })``;
