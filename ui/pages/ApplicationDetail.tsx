import { CircularProgress } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import { useHistory } from "react-router-dom";
import styled from "styled-components";
import Alert from "../components/Alert";
import Button from "../components/Button";
import CommitsTable from "../components/CommitsTable";
import ConditionsTable from "../components/ConditionsTable";
import ErrorPage from "../components/ErrorPage";
import Flex from "../components/Flex";
import GithubDeviceAuthModal from "../components/GithubDeviceAuthModal";
import KeyValueTable from "../components/KeyValueTable";
import LoadingPage from "../components/LoadingPage";
import Modal from "../components/Modal";
import Page from "../components/Page";
import ReconciliationGraph from "../components/ReconciliationGraph";
import Spacer from "../components/Spacer";
import { AppContext } from "../contexts/AppContext";
import { useRequestState } from "../hooks/common";
import {
  AutomationKind,
  GetApplicationResponse,
  RemoveApplicationResponse,
  UnstructuredObject,
} from "../lib/api/applications/applications.pb";
import { getChildren } from "../lib/graph";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
  name: string;
};

function ApplicationDetail({ className, name }: Props) {
  const { applicationsClient, linkResolver } = React.useContext(AppContext);
  const [authSuccess, setAuthSuccess] = React.useState(false);
  const [githubAuthModalOpen, setGithubAuthModalOpen] = React.useState(false);
  const [removeAppModalOpen, setRemoveAppModalOpen] = React.useState(false);
  const [reconciledObjects, setReconciledObjects] = React.useState<
    UnstructuredObject[]
  >([]);
  const [res, loading, error, req] = useRequestState<GetApplicationResponse>();
  const [removeRes, removeLoading, removeError, removeRequest] =
    useRequestState<RemoveApplicationResponse>();
  //for redirects
  const history = useHistory();

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

  React.useEffect(() => {
    if (!removeRes) return;
    //if app is succesfully removed, redirect to applications page
    history.push(linkResolver(PageRoute.Applications));
  }, [removeRes]);

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
      topRight={
        <Button
          color="secondary"
          variant="contained"
          onClick={() => setRemoveAppModalOpen(true)}
        >
          Remove App
        </Button>
      }
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
      <Modal
        //confirm modal for app removal
        open={removeAppModalOpen}
        onClose={() => setRemoveAppModalOpen(false)}
        title="Are You Sure?"
        description={`You are about to remove ${application.name} from Weave GitOps`}
      >
        <Flex align column center wide>
          {authSuccess ? (
            <Flex align>
              <Alert severity="success" message="Authentication Successful" />
            </Flex>
          ) : (
            <>
              <Flex center>
                <Spacer padding="small">
                  <Alert
                    severity="error"
                    title="You are not Authenticated!"
                    message="To remove this app, please authenticate with GitHub"
                  />
                </Spacer>
              </Flex>
              <Flex center>
                <Spacer padding="small">
                  <Button
                    color="secondary"
                    variant="contained"
                    onClick={() => {
                      setGithubAuthModalOpen(true);
                    }}
                  >
                    Authenticate with GitHub
                  </Button>
                </Spacer>
              </Flex>
            </>
          )}
          {removeError && authSuccess && (
            <Flex align center wide>
              <Spacer padding="small">
                <Alert
                  severity="error"
                  title="Error removing Application"
                  message={removeError?.message}
                />
              </Spacer>
            </Flex>
          )}
          {authSuccess && (
            <Flex align center wide>
              <Spacer padding="medium">
                <Button
                  color="secondary"
                  variant="contained"
                  onClick={() =>
                    removeRequest(
                      applicationsClient.RemoveApplication({
                        name: application.name,
                        namespace: application.namespace,
                        //CAN'T FIND AUTOMERGE IN APPLICATION OBJECT
                        autoMerge: true,
                      })
                    )
                  }
                >
                  {removeLoading ? (
                    <CircularProgress color="inherit" size="75%" />
                  ) : (
                    `Delete ${application.name}`
                  )}
                </Button>
              </Spacer>
            </Flex>
          )}
        </Flex>
      </Modal>
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
