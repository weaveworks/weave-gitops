import { FormControl, FormHelperText, MenuItem } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { GithubAuthContext } from "../contexts/GithubAuthContext";
import { useCreateDeployKey } from "../hooks/apps";
import { useIsAuthenticated } from "../hooks/auth";
import { GitProvider } from "../lib/api/applications/applications.pb";
import { WeGONamespace } from "../lib/types";
import { notifyError, notifySuccess } from "../lib/utils";
import Button, { Props as ButtonProps } from "./Button";
import Flex from "./Flex";
import Form from "./Form";
import FormCheckbox from "./FormCheckbox";
import FormInput, { Label } from "./FormInput";
import FormSelect from "./FormSelect";
import Input from "./Input";
import Spacer from "./Spacer";
import Text from "./Text";

type Props = {
  className?: string;
  onSubmit: (state: GitRepoFormState) => void;
};

const initialState = () => ({
  provider: "",
  name: "",
  namespace: WeGONamespace,
  publicRepo: "true",
  intervalMinutes: 3,
  intervalSeconds: 0,
  branch: "main",
  url: "",
});

export type GitRepoFormState = ReturnType<typeof initialState>;

const IntervalInput = styled(FormInput)`
  input {
    width: 40px;
  }
`;

const OAuthButton = ({ provider, ...props }) => (
  <Button {...props}>
    {props.disabled
      ? "Select a Git Provider above to use OAuth"
      : `${provider} OAuth`}
  </Button>
);

type DeployButtonProps = { state: GitRepoFormState } & ButtonProps;

const DeployKeyButton = ({ state, ...props }: DeployButtonProps) => (
  <>
    <Button disabled={!state?.url} {...props}>
      Create Deploy Key Secret
    </Button>
    {!state?.url && (
      <FormHelperText error>
        Populate the Repo URL field above to create a deploy key
      </FormHelperText>
    )}
  </>
);

function AddGitRepoForm({ className, onSubmit }: Props) {
  const { setDialogState, dialogState } = React.useContext(GithubAuthContext);
  const [formState, setFormState] = React.useState<GitRepoFormState>(
    initialState()
  );
  const [secretRef, setSecretRef] = React.useState<string>("");
  const { isAuthenticated, req } = useIsAuthenticated();
  const mutation = useCreateDeployKey();

  const handleChange = (state: { values: GitRepoFormState }) => {
    if (state.values) {
      setFormState(state.values);
    }
  };

  const handleAuthClick = (provider: GitProvider) => {
    if (provider === GitProvider.GitHub) {
      setDialogState(true, formState.name);
    }
  };

  const handleDeployKeyClick = () => {
    mutation
      .mutateAsync({
        secretName: formState.name,
        namespace: formState.namespace,
        provider: formState.provider as GitProvider,
        repoUrl: formState.url,
      })
      .then((res) => {
        notifySuccess("Deploy key secret added succesfully!");
        setSecretRef(res.secretName);
      })
      .catch((err) => notifyError(err.message));
  };

  React.useEffect(() => {
    if (!formState) {
      return;
    }

    if (formState?.provider) {
      req(formState.provider as GitProvider);
    }
  }, [formState.provider, dialogState]);

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
          name="url"
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
          {isAuthenticated ? (
            <DeployKeyButton
              state={formState}
              loading={mutation.isLoading}
              onClick={handleDeployKeyClick}
            />
          ) : (
            <OAuthButton
              disabled={noGitProvider}
              provider={formState?.provider}
              onClick={() => handleAuthClick(formState.provider as GitProvider)}
              type="button"
            />
          )}

          <Spacer m={["medium"]}>
            <FormControl>
              {/* Escape the <Form /> component here to set our own value on the secret ref */}
              {/* We do this to populate the field with the secret we just created via the oauth button. */}
              <Label>Secret Ref</Label>
              <Input
                onChange={(ev) => setSecretRef(ev.target.value)}
                variant="outlined"
                value={secretRef}
              />
              <FormHelperText>
                Reference a secret already on the cluster or populate a secret
                via OAuth
              </FormHelperText>
            </FormControl>
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
})`
  #url {
    min-width: 360px;
  }
`;
