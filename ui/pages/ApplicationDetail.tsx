import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Alert from "../components/Alert";
import CommitsTable from "../components/CommitsTable";
import ConditionsTable from "../components/ConditionsTable";
import ErrorPage from "../components/ErrorPage";
import GithubDeviceAuthModal from "../components/GithubDeviceAuthModal";
import KeyValueTable from "../components/KeyValueTable";
import LoadingPage from "../components/LoadingPage";
import Page from "../components/Page";
import ReconciliationGraph from "../components/ReconciliationGraph";
import { AppContext } from "../contexts/AppContext";
import { useRequestState } from "../hooks/common";
import {
  AutomationKind,
  GetApplicationResponse,
  UnstructuredObject,
} from "../lib/api/applications/applications.pb";
import { getChildren } from "../lib/graph";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
  name: string;
};

function ApplicationDetail({ className, name }: Props) {
  const { applicationsClient } = React.useContext(AppContext);
  const [authSuccess, setAuthSuccess] = React.useState(false);
  const [githubAuthModalOpen, setGithubAuthModalOpen] = React.useState(false);
  const [reconciledObjects, setReconciledObjects] = React.useState<
    UnstructuredObject[]
  >([]);
  const [res, loading, error, req] = useRequestState<GetApplicationResponse>();

  React.useEffect(() => {
    req(applicationsClient.GetApplication({ name, namespace: "wego-system" }));
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
      loading={loading}
      breadcrumbs={[{ page: PageRoute.Applications }]}
      title={name}
      className={className}
    >
      {authSuccess && (
        <Alert severity="success" message="Authentication Successful" />
      )}
      {error && (
        <Alert
          severity="error"
          title="Error fetching Application"
          message={error.message}
        />
      )}
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
        // Get CommitsTable to retry after auth
        app={application}
        onAuthClick={() => setGithubAuthModalOpen(true)}
      />
      <GithubDeviceAuthModal
        onSuccess={() => {
          setAuthSuccess(true);
        }}
        repoName={application.url}
        onClose={() => setGithubAuthModalOpen(false)}
        open={githubAuthModalOpen}
      />
    </Page>
  );
}

export default styled(ApplicationDetail)``;
