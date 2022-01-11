import { createHashHistory } from "history";
import qs from "qs";
import * as React from "react";
import { Redirect, Route, Router, Switch } from "react-router-dom";
import styled from "styled-components";
import AddAppForm, { FormState } from "../../../components/AddAppForm";
import AddKustomizationForm from "../../../components/AddKustomizationForm";
import Page from "../../../components/Page";
import { useCreateApp } from "../../../hooks/apps";
import { formatURL } from "../../../lib/utils";

type Props = {
  className?: string;
};

enum WizardRoute {
  AppForm = "/app_form",
  KustomizationForm = "/kustomization_form",
  SourceForm = "/source_form",
}

type WizardState = {
  appName: string;
  currentStep: WizardRoute;
};

enum ActionType {
  appCreated,
  automationCreated,
}

function pageTitleText(route: string, appName: string) {
  switch (route) {
    case WizardRoute.AppForm:
      return `Add Application`;

    case WizardRoute.KustomizationForm:
      return `Add Flux Kustomization for ${appName}`;

    default:
      break;
  }
}

type Action =
  | { type: ActionType.appCreated; name: string }
  | { type: ActionType.automationCreated };

function reducer(state: WizardState, action: Action) {
  switch (action.type) {
    case ActionType.appCreated:
      return {
        ...state,
        appName: action.name,
        currentStep: WizardRoute.KustomizationForm,
      };
    case ActionType.automationCreated:
      return {
        ...state,
        currentStep: WizardRoute.SourceForm,
      };

    default:
      break;
  }
}
const history = createHashHistory();

function Wizard({ className }: Props) {
  const [state, dispatch] = React.useReducer(reducer, {
    appName: "",
    currentStep: WizardRoute.AppForm,
  });
  const params = qs.parse(history.location.search, { ignoreQueryPrefix: true });
  console.log(params);
  const mutation = useCreateApp();

  const handleAppSubmit = (vals: FormState) => {
    dispatch({ type: ActionType.appCreated, name: vals.name });
    mutation.mutate(vals);
  };

  React.useEffect(() => {
    if (mutation.isSuccess) {
      const u = formatURL(WizardRoute.KustomizationForm, {
        appName: state.appName,
      });
      history.push(u);
    }
  }, [mutation.isSuccess]);

  return (
    <Page
      title={pageTitleText(history.location.pathname, params.appName as string)}
      error={null}
      className={className}
    >
      <Router history={history}>
        <Switch>
          <Route
            exact
            path={WizardRoute.AppForm}
            component={() => (
              <AddAppForm
                loading={mutation.isLoading}
                onSubmit={(state) => handleAppSubmit(state)}
              />
            )}
          />
          <Route
            exact
            path={WizardRoute.KustomizationForm}
            component={AddKustomizationForm}
          />
          <Redirect exact from="/" to={WizardRoute.AppForm} />
        </Switch>
      </Router>
    </Page>
  );
}

export default styled(Wizard).attrs({ className: Wizard.name })``;
