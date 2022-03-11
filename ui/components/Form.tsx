import * as React from "react";
import styled from "styled-components";
import ControlledForm from "./ControlledForm";

interface Props {
  className?: string;
  children?: any;
  onSubmit?: (state: any) => any;
  onChange?: (state: any) => any;
  initialState: any;
}

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
    <ControlledForm
      className={className}
      state={state}
      onChange={(name, value) =>
        dispatch({ type: ActionTypes.setValue, name, value })
      }
      onSubmit={() => onSubmit(state.values)}
    >
      {children}
    </ControlledForm>
  );
}

export default styled(Form).attrs({ className: Form.name })``;
