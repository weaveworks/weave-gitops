import {
  CircularProgress,
  FormControlLabel,
  FormGroup,
  FormHelperText,
  Switch,
  TextField,
} from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Alert from "../components/Alert";
import Button from "../components/Button";
import Flex from "../components/Flex";
import Link from "../components/Link";
import Page from "../components/Page";
import { AppContext } from "../contexts/AppContext";
import { useRequestState } from "../hooks/common";
import { AddApplicationResponse } from "../lib/api/applications/applications.pb";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
};

function isHTTP(uri) {
  return uri.includes("http") || uri.includes("https");
}

function convertGitURLToGitProvider(uri: string) {
  if (isHTTP(uri)) {
    return uri;
  }

  const matches = uri.match(/git@(.*)[/|:](.*)\/(.*)/);
  if (!matches) {
    throw new Error(`could not parse url "${uri}"`);
  }
  const [, provider, org, repo] = matches;

  return `https://${provider}/${org}/${repo}`;
}

const SuccessMessage = styled(
  ({
    link,
    className,
    autoMerged,
  }: {
    className?: string;
    link: string;
    autoMerged: boolean;
  }) => {
    return (
      <div className={className}>
        <div>
          <Alert severity="success" title="Application added successfully!" />
        </div>

        <div className="pr-info">
          <Alert
            severity="info"
            title={autoMerged ? "Application added" : "Pull Request Created"}
            message={
              <div>
                <div>
                  {autoMerged
                    ? `Your application was added to your cluster. 
                          It may take a minute for the application to complete reconciliation and appear on the cluster.`
                    : `A pull request was created for this
                  application. To add the application to your cluster, review
                  and merge the pull request.`}
                </div>
                <Flex wide center>
                  {autoMerged ? (
                    <Link to={PageRoute.Applications}>
                      <Button color="primary" variant="outlined">
                        View Applications
                      </Button>
                    </Link>
                  ) : (
                    <a target="_blank" href={link}>
                      <Button color="primary" variant="outlined">
                        View Open Pull requests
                      </Button>
                    </a>
                  )}
                </Flex>
              </div>
            }
          />
        </div>
      </div>
    );
  }
)`
  a {
    text-decoration: none;
  }

  ${Button} {
    margin-top: 8px;
    margin-left: 8px;
  }

  .pr-info {
    padding-top: 16px;
  }

  .MuiAlert-message {
    width: 100%;
  }
`;

const FormElement = styled.div`
  padding-bottom: 16px;
`;

function AddApplication({ className }: Props) {
  const { applicationsClient } = React.useContext(AppContext);
  const [formState, setFormState] = React.useState({
    name: "",
    namespace: "wego-system",
    path: "",
    branch: "main",
    url: "",
    configUrl: "",
    autoMerge: false,
  });
  const [addRes, loading, error, req] =
    useRequestState<AddApplicationResponse>();
  const [prLink, setPrLink] = React.useState("");

  const handleSubmit = () => {
    req(
      applicationsClient.AddApplication({
        ...formState,
      })
    );
  };

  React.useEffect(() => {
    if (!addRes) {
      return;
    }
    const repoURL = convertGitURLToGitProvider(
      formState.configUrl || formState.url
    );

    setPrLink(`${repoURL.replace(".git", "")}/pulls`);
  }, [addRes]);

  return (
    <Page className={className} title="Add Application">
      {error && (
        <Alert severity="error" title="Error!" message={error.message} />
      )}
      {addRes && addRes.success ? (
        <SuccessMessage autoMerged={formState.autoMerge} link={prLink} />
      ) : (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSubmit();
          }}
        >
          <FormElement>
            <TextField
              onChange={(e) => {
                setFormState({
                  ...formState,
                  name: e.currentTarget.value,
                });
              }}
              required
              id="name"
              label="Name"
              variant="standard"
              value={formState.name}
            />
            <FormHelperText>The name of the application</FormHelperText>
          </FormElement>
          <FormElement>
            <TextField
              onChange={(e) => {
                setFormState({
                  ...formState,
                  namespace: e.currentTarget.value,
                });
              }}
              required
              id="namespace"
              label="Kubernetes Namespace"
              variant="standard"
              value={formState.namespace}
            />
            <FormHelperText>
              The the target namespace for the application
            </FormHelperText>
          </FormElement>
          <FormElement>
            <TextField
              onChange={(e) => {
                setFormState({
                  ...formState,
                  url: e.currentTarget.value,
                });
              }}
              required
              id="url"
              label="Source Repo URL"
              variant="standard"
              value={formState.url}
            />
            <FormHelperText>
              The git repository URL where the application YAML files are stored
            </FormHelperText>
          </FormElement>
          <FormElement>
            <TextField
              onChange={(e) => {
                setFormState({
                  ...formState,
                  configUrl: e.currentTarget.value,
                });
              }}
              id="configUrl"
              label="Config Repo URL"
              variant="standard"
              value={formState.configUrl}
            />
            <FormHelperText>
              The git repository URL to which Weave GitOps will write the GitOps
              Automation files
            </FormHelperText>
          </FormElement>
          <FormElement>
            <TextField
              onChange={(e) => {
                setFormState({
                  ...formState,
                  path: e.currentTarget.value,
                });
              }}
              required
              id="path"
              label="Path"
              variant="standard"
              value={formState.path}
            />
            <FormHelperText>
              The path within the git repository where your application YAMLs
              are stored
            </FormHelperText>
          </FormElement>
          <FormElement>
            <TextField
              onChange={(e) => {
                setFormState({
                  ...formState,
                  branch: e.currentTarget.value,
                });
              }}
              required
              id="branch"
              label="Branch"
              variant="standard"
              value={formState.branch}
            />
            <FormHelperText>
              The git branch to use when reading the application YAMLs
            </FormHelperText>
          </FormElement>
          <FormElement>
            <FormGroup>
              <FormControlLabel
                control={
                  <Switch
                    onChange={(e) =>
                      setFormState({
                        ...formState,
                        autoMerge: e.currentTarget.checked,
                      })
                    }
                    checked={formState.autoMerge}
                  />
                }
                label="Auto Merge"
              />
            </FormGroup>
            <FormHelperText>
              If checked, Weave GitOps will automatically merge the application
              into the default branch instead of doing a pull request
            </FormHelperText>
          </FormElement>
          <Flex wide end>
            {loading ? (
              <CircularProgress />
            ) : (
              <Button variant="contained" color="primary" type="submit">
                Submit
              </Button>
            )}
          </Flex>
        </form>
      )}
    </Page>
  );
}

export default styled(AddApplication).attrs({
  className: AddApplication.name,
})`
  h2 {
    color: ${(props) => props.theme.colors.black};
  }

  .MuiFormHelperText-root {
    color: ${(props) => props.theme.colors.black};
  }
`;
