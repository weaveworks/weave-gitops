import * as React from "react";
import styled from "styled-components";
import Button from "../../../components/Button";
import DataTable from "../../../components/DataTable";
import FancyCard from "../../../components/FancyCard";
import Link from "../../../components/Link";
import Page from "../../../components/Page";
import Spacer from "../../../components/Spacer";
import Text from "../../../components/Text";
import { AppContext } from "../../../contexts/AppContext";
import { useGetApplication, useRemoveApp } from "../../../hooks/apps";
import { useGetKustomizations } from "../../../hooks/kustomizations";
import { V2Routes, WeGONamespace } from "../../../lib/types";
import { formatURL } from "../../../lib/utils";
import EmptyApplication from "./EmptyApplication";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function Kustomizations({ kustomizations }) {
  return (
    <>
      <div>
        <h3>Kustomizations</h3>
      </div>
      <DataTable
        sortFields={["name"]}
        fields={[
          {
            label: "Name",
            value: (k) => (
              <Link
                to={formatURL(V2Routes.Kustomization, {
                  name: k.name,
                  namespace: k.namespace,
                })}
              >
                {k.name}
              </Link>
            ),
          },
          {
            label: "Namespace",
            value: "namespace",
          },
        ]}
        rows={kustomizations}
      />
    </>
  );
}

function Application({ className, name, namespace = WeGONamespace }: Props) {
  const { navigate } = React.useContext(AppContext);
  const { data } = useGetApplication(name);
  const {
    data: kustomizationRes,
    isLoading,
    error,
  } = useGetKustomizations(name, namespace);
  const remove = useRemoveApp();

  const handleRemove = () => {
    remove.mutateAsync({ name: data?.app.name || name }).then(() => {
      navigate.internal(V2Routes.ApplicationList);
    });
  };

  const kustomizations = kustomizationRes?.kustomizations;

  return (
    <Page
      className={className}
      error={error}
      loading={isLoading}
      title={name}
      actions={
        <Button onClick={handleRemove} color="secondary">
          Remove App
        </Button>
      }
    >
      <div>
        <Text>{data?.app.description}</Text>
      </div>

      <Spacer m={["small"]}>
        {kustomizations?.length > 0 ? (
          <Kustomizations kustomizations={kustomizations} />
        ) : (
          <EmptyApplication appName={name} />
        )}
      </Spacer>
    </Page>
  );
}

export default styled(Application).attrs({ className: Application.name })`
  ${FancyCard} {
    max-width: 272px;
  }
`;
