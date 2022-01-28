import * as React from "react";
import styled from "styled-components";
import AddAppForm from "../../components/AddAppForm";
import Page from "../../components/Page";
import { AppContext } from "../../contexts/AppContext";
import { useCreateApp } from "../../hooks/apps";
import { V2Routes, WeGONamespace } from "../../lib/types";

type Props = {
  className?: string;
};

function NewApp({ className }: Props) {
  const { navigate } = React.useContext(AppContext);

  const mutation = useCreateApp();

  React.useEffect(() => {
    if (mutation.isSuccess) {
      navigate.internal(V2Routes.Application, {
        appName: mutation.data.app.name,
      });
    }
  }, [mutation.isSuccess]);

  return (
    <Page error={mutation.error} title="Add Application" className={className}>
      <AddAppForm
        loading={mutation.isLoading}
        onSubmit={(state) => {
          mutation.mutate({ ...state, namespace: WeGONamespace });
        }}
      />
    </Page>
  );
}

export default styled(NewApp).attrs({ className: NewApp.name })``;
