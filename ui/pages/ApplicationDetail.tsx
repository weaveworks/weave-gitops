import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import ActionBar from "../components/ActionBar";
import Alert from "../components/Alert";
import Button from "../components/Button";
import CommitsTable from "../components/CommitsTable";
import ConditionsTable from "../components/ConditionsTable";
import ErrorPage from "../components/ErrorPage";
import GithubDeviceAuthModal from "../components/GithubDeviceAuthModal";
import KeyValueTable from "../components/KeyValueTable";
import LoadingPage from "../components/LoadingPage";
import Page from "../components/Page";
import ReconciliationGraph from "../components/ReconciliationGraph";
import Spacer from "../components/Spacer";
import { AppContext } from "../contexts/AppContext";
import CallbackStateContextProvider from "../contexts/CallbackStateContext";
import { useRequestState } from "../hooks/common";
import {
  AutomationKind,
  GetApplicationResponse,
  SyncApplicationResponse,
  UnstructuredObject,
} from "../lib/api/applications/applications.pb";
import { getChildren } from "../lib/graph";
import { PageRoute } from "../lib/types";
import { formatURL } from "../lib/utils";

type Props = {
  className?: string;
  name: string;
};

function ApplicationDetail({ className, name }: Props) {
  const { applicationsClient, notifySuccess, navigate } =
    React.useContext(AppContext);
  const [authSuccess, setAuthSuccess] = React.useState(false);
  const [githubAuthModalOpen, setGithubAuthModalOpen] = React.useState(false);
  const [reconciledObjects, setReconciledObjects] = React.useState<
    UnstructuredObject[]
  >([]);
  const [provider, setProvider] = React.useState("");
  const [res, loading, error, req] = useRequestState<GetApplicationResponse>();
  const [syncRes, syncLoading, syncError, syncRequest] =
    useRequestState<SyncApplicationResponse>();
  const { getCallbackState, clearCallbackState } = React.useContext(AppContext);

  const callbackState = getCallbackState();

  if (callbackState) {
    setAuthSuccess(true);
    clearCallbackState();
  }

  React.useEffect(() => {
    const p = async () => {
      const res = await applicationsClient.GetApplication({
        name,
        namespace: "wego-system",
      });

      const { provider } = await applicationsClient.ParseRepoURL({
        url: res.application.url,
      });
      setProvider(provider);

      return { ...res, provider };
    };

    req(p());
  }, [name]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    const uniqKinds = _.uniqBy(res.application.reconciledObjectKinds, "kind");
    if (uniqKinds.length === 0) {
      return;
    }

    getChildren(applicationsClient, res.application, uniqKinds).then((objs) =>
      setReconciledObjects(objs)
    );
  }, [res]);

  React.useEffect(() => {
    if (syncRes) {
      notifySuccess("App Sync Successful");
    }
  }, [syncRes]);

  if (error) {
    return (
      <ErrorPage
        breadcrumbs={[{ page: PageRoute.Applications }]}
        title={name}
        error={{ message: error.message, title: "Error fetching Application" }}
      />
    );
  }

  if ((!res && !error) || loading) {
    return <LoadingPage />;
  }

  const { application = {} } = res;

  return (
    <Page
      loading={loading ? true : false}
      breadcrumbs={[{ page: PageRoute.Applications }]}
      title={name}
      className={className}
    >
      <CallbackStateContextProvider
        callbackState={{
          page: formatURL(PageRoute.ApplicationDetail, { name }),
          state: { authSuccess: false },
        }}
      >
        {syncError ? (
          <Alert
            severity="error"
            title="Error syncing Application"
            message={syncError.message}
          />
        ) : (
          authSuccess && (
            <Alert severity="success" message="Authentication Successful" />
          )
        )}
        <ActionBar>
          <Button
            loading={syncLoading}
            onClick={() => {
              syncRequest(
                applicationsClient.SyncApplication({
                  name: application.name,
                  namespace: application.namespace,
                })
              );
            }}
          >
            Sync App
          </Button>
          <Spacer padding="small" />
          <Button
            color="secondary"
            onClick={() =>
              navigate.internal(PageRoute.ApplicationRemove, { name })
            }
          >
            Remove App
          </Button>
        </ActionBar>
        <KeyValueTable
          columns={4}
          pairs={[
            { key: "Name", value: application.name },
            { key: "Deployment Type", value: application.deploymentType },
            { key: "URL", value: application.url },
            { key: "Path", value: application.path },
          ]}
        />
        <ReconciliationGraph
          objects={reconciledObjects}
          parentObject={application}
          parentObjectKind="Application"
        />
        <h3>Source Conditions</h3>
        <ConditionsTable conditions={application.source?.conditions} />
        <h3>Automation Conditions</h3>
        <ConditionsTable
          conditions={
            application.deploymentType == AutomationKind.Kustomize
              ? application.kustomization?.conditions
              : application.helmRelease?.conditions
          }
        />
        <h3>Commits</h3>
        <CommitsTable
          app={application}
          authSuccess={authSuccess}
          onAuthClick={() => {
            if (provider === "GitHub") setGithubAuthModalOpen(true);
          }}
          provider={provider}
        />
        <GithubDeviceAuthModal
          bodyClassName="auth-modal-size"
          onSuccess={() => {
            setAuthSuccess(true);
          }}
          repoName={application.url}
          onClose={() => {
            setGithubAuthModalOpen(false);
          }}
          open={githubAuthModalOpen}
        />
      </CallbackStateContextProvider>
    </Page>
  );
}

export default styled(ApplicationDetail)``;
