import { Divider, MenuItem } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useListSources } from "../hooks/sources";
import Button from "./Button";
import Flex from "./Flex";
import Form from "./Form";
import FormInput from "./FormInput";
import FormSelect from "./FormSelect";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

type Props = {
  className?: string;
  onCreateSourceClick: (ev: any) => void;
  onSubmit: (state: FormState) => void;
  onChange: (state: FormState) => void;
  initialState: FormState;
  loading: boolean;
};

const defaultInitialState = () => ({
  name: "",
  namespace: "",
  source: "",
  path: "",
});

type FormState = ReturnType<typeof defaultInitialState>;

function AddKustomizationForm({
  className,
  onSubmit,
  onChange,
  onCreateSourceClick,
  initialState,
  loading,
}: Props) {
  const {
    data: { result: sources = [] },
  } = useListSources();

  return (
    <div className={className}>
      <Spacer m={["large"]}>
        <Form
          onChange={onChange}
          onSubmit={onSubmit}
          initialState={initialState}
        >
          <Flex between>
            <FormInput
              name="name"
              label="Name"
              required
              helperText="The name of the kustomization"
            />
            <FormInput
              name="namespace"
              label="Kubernetes Namespace"
              required
              helperText="The namespace where GitOps automation objects will be stored."
            />
          </Flex>
          <Spacer m={["medium"]}>
            <FormSelect
              name="source"
              label="Select Source"
              required
              helperText="The git repository URL where the application YAML files are stored"
            >
              {sources.length > 0 ? (
                _.map(sources, (s, i) => (
                  <MenuItem value={s.name} key={i}>
                    {s.name}
                  </MenuItem>
                ))
              ) : (
                <MenuItem disabled divider>
                  <Text italic>No existing sources found</Text>
                </MenuItem>
              )}

              <Divider />
              <MenuItem value="create">
                <Link onClick={onCreateSourceClick} to="">
                  <Text style={{ textTransform: "uppercase" }}>
                    Create Source
                  </Text>
                </Link>
              </MenuItem>
            </FormSelect>
          </Spacer>
          <Spacer m={["medium"]}>
            <FormInput
              name="path"
              label="Path"
              required
              helperText="The path within the git repository where your application YAMLs are stored"
            />
          </Spacer>
          <Spacer m={["medium"]}>
            <Button loading={loading} type="submit">
              Create Kustomization
            </Button>
          </Spacer>
        </Form>
      </Spacer>
    </div>
  );
}

export default styled(AddKustomizationForm).attrs({
  className: AddKustomizationForm.name,
})``;
