import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Alert from "../components/Alert";
import CommitsTable from "../components/CommitsTable";
import ConditionsTable from "../components/ConditionsTable";
import GithubDeviceAuthModal from "../components/GithubDeviceAuthModal";
import KeyValueTable from "../components/KeyValueTable";
import LoadingPage from "../components/LoadingPage";
import Page from "../components/Page";
import ReconciliationGraph from "../components/ReconciliationGraph";
import { AppContext } from "../contexts/AppContext";
import useApplications from "../hooks/applications";
import {
  Application,
  UnstructuredObject,
  AutomationKind,
} from "../lib/api/applications/applications.pb";
import { getChildren } from "../lib/graph";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
  name: string;
};

function ApplicationDetail({ className, name }: Props) {
  const [app, setApp] = React.useState<Application>({});
  const [authSuccess, setAuthSuccess] = React.useState(false);
  const [githubAuthModalOpen, setGithubAuthModalOpen] = React.useState(false);
  const { applicationsClient, doAsyncError } = React.useContext(AppContext);
  const [reconciledObjects, setReconciledObjects] = React.useState<
    UnstructuredObject[]
  >([]);

  const { getApplication, loading } = useApplications();

  React.useEffect(() => {
    getApplication(name)
      .then((app) => {
        setApp(app as Application);
      })
      .catch((err) =>
        doAsyncError("Error fetching application detail", err.message)
      );
  }, []);

  React.useEffect(() => {
    if (!app) {
      return;
    }

    const uniqKinds = _.uniqBy(app.reconciledObjectKinds, "kind");
    if (uniqKinds.length === 0) {
      return;
    }

    getChildren(applicationsClient, app, uniqKinds).then((objs) =>
      setReconciledObjects(objs)
    );
  }, [app]);

  if (loading) {
    return <LoadingPage />;
  }

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
      <KeyValueTable
        columns={4}
        pairs={[
          { key: "Name", value: app.name },
          { key: "Deployment Type", value: app.deploymentType },
          { key: "URL", value: app.url },
          { key: "Path", value: app.path },
        ]}
      />
      <ReconciliationGraph
        objects={reconciledObjects}
        parentObject={app}
        parentObjectKind="Application"
      />
      <h3>Source Conditions</h3>
      <ConditionsTable conditions={app.source?.conditions} />
      <h3>Automation Conditions</h3>
      <ConditionsTable conditions={app.deploymentType == AutomationKind.Kustomize ? app.kustomization?.conditions : app.helmRelease?.conditions} />

      <h3>Commits</h3>
      <CommitsTable
        app={app}
        onAuthClick={() => setGithubAuthModalOpen(true)}
      />
      <GithubDeviceAuthModal
        onSuccess={() => {
          setAuthSuccess(true);
          // Get CommitsTable to retry after auth
          setApp({ ...app });
        }}
        repoName={app.url}
        onClose={() => setGithubAuthModalOpen(false)}
        open={githubAuthModalOpen}
      />
    </Page>
  );
}

export default styled(ApplicationDetail)``;
