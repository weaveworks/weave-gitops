import * as React from "react";
import { useHistory } from "react-router-dom";
import styled from "styled-components";
import AddKustomizationForm from "../../components/AddKustomizationForm";
import Page from "../../components/Page";
import { useCreateKustomization } from "../../hooks/kustomizations";
import { SourceRefKind } from "../../lib/api/app/source.pb";
import { V2Routes, WeGONamespace } from "../../lib/types";
import { formatURL } from "../../lib/utils";

type Props = {
  className?: string;
  appName: string;
  query?: any;
};

const defaultInitialState = () => ({
  name: "",
  namespace: WeGONamespace,
  source: "",
  path: "./",
});

type FormState = ReturnType<typeof defaultInitialState>;

function stateFromURL(query): FormState {
  const def: FormState = defaultInitialState();
  if (!query || !query.state) {
    return def;
  }
  return query?.state ? JSON.parse(query.state) : def;
}

function AddKustomization({ className, appName, query }: Props) {
  const history = useHistory();
  const formStateFromURL = { ...defaultInitialState(), ...stateFromURL(query) };
  const [formState, setFormState] = React.useState<FormState>(null);
  const mutation = useCreateKustomization();

  const sourceUrl = () =>
    formatURL(V2Routes.AddSource, {
      appName,
      next: V2Routes.NewApp,
      state: JSON.stringify(formState),
    });

  const handleSourceCreateClick = (ev) => {
    // Save form state on navigation
    ev.preventDefault();
    // const vals = _.omit(formState, "source");
    // const q = qs.stringify({
    //   ...formStateFromURL,
    //   state: JSON.stringify(vals),
    // });
    // history.replace({ pathname: history.location.pathname, search: q });
    history.push(sourceUrl());
  };

  const handleSubmit = async (state: FormState) => {
    const namespace = state.namespace || defaultInitialState().namespace;

    await mutation.mutateAsync({
      ...state,
      appName,
      sourceRef: { kind: SourceRefKind.GitRepository, name: state.source },
      namespace,
    });

    if (!mutation.isError) {
      history.push(
        formatURL(V2Routes.Kustomization, { name: state.name, namespace })
      );
    }
  };

  return (
    <Page
      error={mutation.error}
      className={className}
      title={`Add Kustomization${appName ? ` for ${appName}` : ""}`}
    >
      <AddKustomizationForm
        loading={mutation.isLoading}
        initialState={formStateFromURL}
        onSubmit={handleSubmit}
        onChange={setFormState}
        onCreateSourceClick={handleSourceCreateClick}
      />
    </Page>
  );
}

export default styled(AddKustomization).attrs({
  className: AddKustomization.name,
})``;
