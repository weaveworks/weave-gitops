import { MenuItem } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { GitProvider } from "../lib/api/applications/applications.pb";
import { WeGONamespace } from "../lib/types";
import Button from "./Button";
import Flex from "./Flex";
import Form from "./Form";
import FormCheckbox from "./FormCheckbox";
import FormInput, { Label } from "./FormInput";
import FormSelect from "./FormSelect";
import Spacer from "./Spacer";
import Text from "./Text";

type Props = {
  className?: string;
  onSubmit: (state: FormState) => void;
};

const initialState = () => ({
  provider: "",
  name: "",
  namespace: WeGONamespace,
  publicRepo: "true",
  intervalMinutes: 3,
  intervalSeconds: 0,
});

type FormState = ReturnType<typeof initialState>;

const IntervalInput = styled(FormInput)`
  input {
    width: 40px;
  }
`;

function AddGitRepoForm({ className, onSubmit }: Props) {
  const [formState, setFormState] = React.useState<FormState>(null);

  const handleChange = (state: { values: FormState }) => {
    if (state.values) {
      setFormState(state.values);
    }
  };

  const noGitProvider =
    !formState?.provider || formState?.provider === GitProvider.Unknown;

  return (
    <Form
      className={className}
      onSubmit={onSubmit}
      onChange={handleChange}
      initialState={initialState()}
    >
      <FormSelect
        name="provider"
        label="Select Provider"
        required
        helperText=""
      >
        {_.map(GitProvider, (p) => {
          const gp = GitProvider[p];

          return (
            <MenuItem value={gp} key={gp}>
              {/* <ListItemIcon></ListItemIcon> */}
              {gp}
            </MenuItem>
          );
        })}
      </FormSelect>
      <Spacer m={["large"]}>
        <Flex start>
          <FormInput
            name="name"
            label="Name"
            required
            helperText="The desired name of the GitRepository object"
          />
          <Spacer m={["none", "none", "none", "large"]}>
            <FormInput
              name="namespace"
              label="Kubernetes Namespace"
              required
              helperText="The namespace where GitOps source objects will be stored."
            />
          </Spacer>
        </Flex>
      </Spacer>
      <Spacer m={["large", "large"]}>
        <FormInput
          name="repoURL"
          label="Repository URL"
          required
          helperText="The git repository URL. We support https and ssh!"
        />
      </Spacer>
      <Flex start>
        <FormInput
          name="branch"
          label="Branch"
          required
          helperText="The git branch to use when reading the application YAMLs"
        />
        <Spacer m={["none", "none", "none", "large"]}>
          <Flex column>
            <Label>Interval</Label>
            <Flex>
              <Flex align>
                <IntervalInput name="intervalMinutes" label="" helperText="" />
                <Spacer m={["none", "none", "small", "small"]}>
                  <Text>minutes</Text>
                </Spacer>
              </Flex>

              <Flex align>
                <IntervalInput name="intervalSeconds" label="" helperText="" />
                <Spacer m={["none", "none", "small", "small"]}>
                  <Text>seconds</Text>
                </Spacer>
              </Flex>
            </Flex>
          </Flex>
        </Spacer>
      </Flex>
      <hr />
      <Flex align start>
        <FormCheckbox label="" name="publicRepo" />
        Public Repo
      </Flex>
      {!formState?.publicRepo && (
        <Spacer m={["medium"]}>
          <Button disabled={noGitProvider} type="button">
            {noGitProvider
              ? "Select a Git Provider above to use OAuth"
              : `${formState?.provider} OAuth`}
          </Button>

          <Spacer m={["medium"]}>
            <FormInput
              name="secretRef"
              label="Secret Ref"
              required
              helperText="Reference a secret already on the cluster instead of using OAuth"
            />
          </Spacer>
        </Spacer>
      )}

      <Spacer m={["xl"]}>
        <Flex wide end>
          <Button type="submit">Create Git Repository</Button>
        </Flex>
      </Spacer>
    </Form>
  );
}

export default styled(AddGitRepoForm).attrs({
  className: AddGitRepoForm.name,
})``;
