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
import AuthAlert from "../components/AuthAlert";
import Button from "../components/Button";
import Flex from "../components/Flex";
import GithubDeviceAuthModal from "../components/GithubDeviceAuthModal";
import Link from "../components/Link";
import Page from "../components/Page";
import RepoInputWithAuth from "../components/RepoInputWithAuth";
import { AppContext } from "../contexts/AppContext";
import CallbackStateContextProvider from "../contexts/CallbackStateContext";
import { useAddApplication } from "../hooks/applications";
import { GitProvider } from "../lib/api/applications/applications.pb";
import { GrpcErrorCodes, PageRoute } from "../lib/types";

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
                      <Button>View Applications</Button>
                    </Link>
                  ) : (
                    <a target="_blank" href={link}>
                      <Button>View Open Pull requests</Button>
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

  .MuiFormControl-root {
    min-width: 360px;
  }
`;

function AddApplication({ className }: Props) {
  const {
    getCallbackState,
    clearCallbackState,
    getProviderToken,
  } = React.useContext(AppContext);
  const formRef = React.useRef<HTMLFormElement>();

  let initialFormState = {
    name: "",
    namespace: "wego-system",
    path: "./",
    branch: "main",
    url: "",
    configRepo: "",
    autoMerge: false,
    provider: null,
  };

  const callbackState = getCallbackState();

  if (callbackState) {
    initialFormState = {
      ...initialFormState,
      ...callbackState.state,
    };
    // Once we read it into the form, clear the state to avoid it auto-populating all the time
    clearCallbackState();
  }

  const [formState, setFormState] = React.useState(initialFormState);
  const [addRes, loading, error, req] = useAddApplication();
  const [prLink, setPrLink] = React.useState("");
  const [authOpen, setAuthOpen] = React.useState(false);
  const [authSuccess, setAuthSuccess] = React.useState(false);

  const handleSubmit = () => {
    req(formState.provider, formState);
  };

  const handleAuthSuccess = () => {
    setAuthSuccess(true);
  };

  const handleAuthClick = () => {
    setAuthOpen(true);
  };

  React.useEffect(() => {
    if (!addRes) {
      return;
    }
    const repoURL = convertGitURLToGitProvider(
      formState.configRepo || formState.url
    );

    setPrLink(`${repoURL.replace(".git", "")}/pulls`);
  }, [addRes]);

  const credentialsDetected =
    authSuccess || !!getProviderToken(formState.provider) || !!callbackState;

  return (
    <Page className={className} title="Add Application">
      <CallbackStateContextProvider
        callbackState={{ page: PageRoute.ApplicationAdd, state: formState }}
      >
        {!authSuccess &&
          error &&
          error.code === GrpcErrorCodes.Unauthenticated && (
            <AuthAlert
              provider={formState.provider}
              title="Error adding application"
              onClick={handleAuthClick}
            />
          )}
        {error && error.code !== GrpcErrorCodes.Unauthenticated && (
          <Alert
            severity="error"
            title="Error adding application"
            message={error.message}
          />
        )}
        {addRes && addRes.success ? (
          <SuccessMessage autoMerged={formState.autoMerge} link={prLink} />
        ) : (
          <form
            ref={formRef}
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
                label="Weave GitOps Kubernetes Namespace"
                variant="standard"
                value={formState.namespace}
              />
              <FormHelperText>
                The namespace where GitOps automation objects will be stored.
              </FormHelperText>
            </FormElement>
            <FormElement>
              <RepoInputWithAuth
                isAuthenticated={!!formState.url && credentialsDetected}
                onChange={(e) => {
                  setFormState({
                    ...formState,
                    url: e.currentTarget.value,
                  });
                }}
                onProviderChange={(provider: GitProvider) => {
                  setFormState({ ...formState, provider });
                }}
                onAuthClick={(provider) => {
                  if (provider === GitProvider.GitHub) {
                    setAuthOpen(true);
                  }
                }}
                required
                id="url"
                label="Source Repo URL"
                variant="standard"
                value={formState.url}
                helperText="The git repository URL where the application YAML files are stored"
              />
            </FormElement>
            <FormElement>
              <TextField
                onChange={(e) => {
                  setFormState({
                    ...formState,
                    configRepo: e.currentTarget.value,
                  });
                }}
                id="configRepo"
                label="Config Repo URL"
                variant="standard"
                value={formState.configRepo}
              />
              <FormHelperText>
                The git repository URL to which Weave GitOps will write the
                GitOps Automation files. Defaults to the Source Repo URL.
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
                      color="primary"
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
                If checked, Weave GitOps will automatically merge the
                application into the default branch instead of doing a pull
                request
              </FormHelperText>
            </FormElement>
            <Flex wide end>
              {loading ? (
                <CircularProgress />
              ) : (
                <Button type="submit">Submit</Button>
              )}
            </Flex>
          </form>
        )}
        <GithubDeviceAuthModal
          onSuccess={handleAuthSuccess}
          onClose={() => setAuthOpen(false)}
          open={authOpen}
          repoName={formState.url}
        />
      </CallbackStateContextProvider>
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

  .MuiFormControl-root {
    width: 420px;
  }
`;
