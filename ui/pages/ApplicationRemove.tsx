import {
  CircularProgress,
  FormControlLabel,
  FormHelperText,
  Switch,
} from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { GithubDeviceAuthModal } from "..";
import Alert from "../components/Alert";
import AuthAlert from "../components/AuthAlert";
import Button from "../components/Button";
import Flex from "../components/Flex";
import Page from "../components/Page";
import Spacer from "../components/Spacer";
import Text from "../components/Text";
import { AppContext } from "../contexts/AppContext";
import CallbackStateContextProvider from "../contexts/CallbackStateContext";
import { useAppRemove } from "../hooks/applications";
import { GitProvider } from "../lib/api/applications/applications.pb";
import { GrpcErrorCodes, PageRoute } from "../lib/types";
import { convertGitURLToGitProvider, formatURL, poller } from "../lib/utils";

type Props = {
  className?: string;
  name: string;
};

const RepoRemoveStatus = ({
  done,
  autoMerge,
  url,
  provider,
}: {
  done: boolean;
  autoMerge: boolean;
  url: string;
  provider: GitProvider;
}) =>
  done ? (
    <Alert
      severity="info"
      title="Removed from Git Repo"
      message={
        autoMerge ? (
          "The application successfully removed from your git repository"
        ) : (
          <Flex align>
            <p>A PR has been successfully created </p>
            <Spacer padding="xs" />
            <a
              target="_blank"
              href={`${convertGitURLToGitProvider(url).replace(".git", "")}/${
                provider === GitProvider.GitLab ? "-/merge_requests" : "pulls"
              }`}
            >
              <Button>View Open Pull Requests</Button>
            </a>
          </Flex>
        )
      }
    />
  ) : null;

const ClusterRemoveStatus = ({ done }: { done: boolean }) =>
  done ? (
    <Alert
      severity="success"
      title="Removed from cluster"
      message="The application was removed from your cluster"
    />
  ) : (
    <Flex wide center align>
      <CircularProgress />
      <Spacer margin="small" />
      <div>Removing from cluster...</div>
    </Flex>
  );

const Prompt = ({
  onRemove,
  name,
  autoMerge,
  setAutoMerge,
}: {
  name: string;
  onRemove: () => void;
  autoMerge: boolean;
  setAutoMerge: any;
}) => (
  <Flex column center>
    <Flex wide center>
      <Text size="large" bold>
        Are you sure you want to remove the application {name}?
      </Text>
    </Flex>
    <Flex wide center>
      <Spacer padding="small">
        Removing this application will remove any Kubernetes objects that were
        created by the application
      </Spacer>
    </Flex>
    <Flex wide align center column>
      <FormControlLabel
        control={
          <Switch
            color="primary"
            onChange={() => setAutoMerge(!autoMerge)}
            checked={autoMerge}
          />
        }
        label="Auto Merge"
      />
      <FormHelperText>
        If checked, Weave GitOps will automatically remove the application from
        the default branch instead of doing a pull request
      </FormHelperText>
    </Flex>
    <Flex wide center>
      <Spacer padding="small">
        <Button onClick={onRemove} variant="contained" color="secondary">
          Remove {name}
        </Button>
      </Spacer>
    </Flex>
  </Flex>
);

function ApplicationRemove({ className, name }: Props) {
  const { applicationsClient } = React.useContext(AppContext);
  const [application, setApplication] = React.useState(null);
  const [autoMerge, setAutoMerge] = React.useState(false);
  const [repoRemoveRes, repoRemoving, error, remove] = useAppRemove();
  const [repoInfo, setRepoInfo] = React.useState({
    provider: null,
    repoName: null,
  });
  const [removedFromCluster, setRemovedFromCluster] = React.useState(false);
  const [authOpen, setAuthOpen] = React.useState(false);
  const [authSuccess, setAuthSuccess] = React.useState(false);
  const [appError, setAppError] = React.useState(null);
  const { getCallbackState, clearCallbackState } = React.useContext(AppContext);

  const callbackState = getCallbackState();

  if (callbackState) {
    setAuthSuccess(true);
    clearCallbackState();
  }

  React.useEffect(() => {
    (async () => {
      try {
        const { application } = await applicationsClient.GetApplication({
          name,
          namespace: "wego-system",
        });

        setApplication(application);

        const { provider, name: repoName } =
          await applicationsClient.ParseRepoURL({ url: application.url });

        setRepoInfo({ provider, repoName });
      } catch (e) {
        setAppError(e.message);
      }
    })();
  }, [name]);

  React.useEffect(() => {
    if (!repoRemoveRes) return;
    const poll = poller(() => {
      applicationsClient
        .GetApplication({ name, namespace: "wego-system" })
        .catch((e) => {
          clearInterval(poll);
          // Once we get a 404, the app is gone for good
          if (e.code === GrpcErrorCodes.NotFound)
            return setRemovedFromCluster(true);
        });
    }, 5000);

    return () => {
      clearInterval(poll);
    };
  }, [repoRemoveRes]);

  const handleRemoveClick = () => {
    remove(repoInfo.provider, {
      name,
      namespace: "wego-system",
      autoMerge: autoMerge,
    });
  };

  const handleAuthSuccess = () => {
    setAuthSuccess(true);
  };

  if (!repoInfo) {
    return <CircularProgress />;
  }

  if (appError) {
    return (
      <Page className={className}>
        <Alert severity="error" title="Error" message={appError} />
      </Page>
    );
  }

  console.log(application);

  return (
    <Page className={className}>
      <CallbackStateContextProvider
        callbackState={{
          page: formatURL(PageRoute.ApplicationRemove, { name }),
          state: { authSuccess: false },
        }}
      >
        {!authSuccess &&
          error &&
          error.code === GrpcErrorCodes.Unauthenticated && (
            <AuthAlert
              title="Error"
              provider={repoInfo.provider}
              onClick={() => setAuthOpen(true)}
            />
          )}
        {error && error.code !== GrpcErrorCodes.Unauthenticated && (
          <Alert title="Error" severity="error" message={error?.message} />
        )}

        {repoRemoving && (
          <Flex wide center align>
            <CircularProgress />
            <Spacer margin="small" />
            <div>Remove operation in progress...</div>
          </Flex>
        )}
        {!repoRemoveRes && !repoRemoving && !removedFromCluster && (
          <Prompt
            name={name}
            onRemove={handleRemoveClick}
            autoMerge={autoMerge}
            setAutoMerge={setAutoMerge}
          />
        )}
        {(repoRemoving || repoRemoveRes) && (
          <RepoRemoveStatus
            done={!repoRemoving}
            url={application.url}
            autoMerge={autoMerge}
            provider={repoInfo.provider}
          />
        )}
        <Spacer margin="small" />
        {repoRemoveRes && <ClusterRemoveStatus done={removedFromCluster} />}
        <GithubDeviceAuthModal
          onSuccess={handleAuthSuccess}
          onClose={() => setAuthOpen(false)}
          open={authOpen}
          repoName={name}
        />
      </CallbackStateContextProvider>
    </Page>
  );
}

export default styled(ApplicationRemove).attrs({
  className: ApplicationRemove.name,
})``;
