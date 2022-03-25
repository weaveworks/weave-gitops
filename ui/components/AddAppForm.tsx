import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import Flex from "./Flex";
import Form from "./Form";
import FormInput from "./FormInput";
import Spacer from "./Spacer";

type Props = {
  className?: string;
  loading?: boolean;
  initialState?: FormState;
  onSubmit: (state: FormState) => void;
};

export type FormState = ReturnType<typeof defaultState>;

const defaultState = () => ({
  name: "",
  displayName: "",
  description: "",
});

function AddAppForm({
  className,
  onSubmit,
  loading,
  initialState = defaultState(),
}: Props) {
  return (
    <Form className={className} initialState={initialState} onSubmit={onSubmit}>
      <Flex start align between>
        <Spacer m={["none", "large"]}>
          <FormInput
            helperText="The name of the application.  Must be unique"
            label="Name"
            name="name"
            required
            variant="outlined"
          />
        </Spacer>
        <Spacer>
          <FormInput
            helperText="The name of the application to be displayed in Weave GitOps."
            label="Display Name"
            name="displayName"
            variant="outlined"
          />
        </Spacer>
      </Flex>

      <Flex wide start>
        <FormInput
          helperText="Tell your future self what this app is for"
          label="Description"
          multiline
          name="description"
          rows={4}
          variant="outlined"
        />
      </Flex>

      <Spacer m={["large"]}>
        <Button loading={loading} color="primary" type="submit">
          Submit
        </Button>
      </Spacer>
    </Form>
  );
}

export default styled(AddAppForm).attrs({ className: AddAppForm.name })`
  ${FormInput} input {
    min-width: 300px;
  }

  ${FormInput} {
    width: 100%;

    .MuiFormControl-root,
    textarea {
      box-sizing: border-box;
      width: 100%;
    }
  }
`;
