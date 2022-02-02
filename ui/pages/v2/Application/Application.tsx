import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import AddSourceButton from "../../../components/AddSourceButton";
import AutomationsTable from "../../../components/AutomationsTable";
import Button from "../../../components/Button";
import FancyCard from "../../../components/FancyCard";
import Flex from "../../../components/Flex";
import Page from "../../../components/Page";
import SourcesTable from "../../../components/SourcesTable";
import Spacer from "../../../components/Spacer";
import Text from "../../../components/Text";
import { AppContext } from "../../../contexts/AppContext";
import { useGetApplication, useRemoveApp } from "../../../hooks/apps";
import { useGetKustomizations } from "../../../hooks/kustomizations";
import { useListSources } from "../../../hooks/sources";
import { AutomationKind } from "../../../lib/api/applications/applications.pb";
import { V2Routes, WeGONamespace } from "../../../lib/types";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function Application({ className, name, namespace = WeGONamespace }: Props) {
  const { navigate } = React.useContext(AppContext);
  const { data, error } = useGetApplication(name);
  const {
    data: kustomizationRes,
    isLoading,
    error: kustErr,
  } = useGetKustomizations(name, namespace);
  const { data: sources, error: sourcesErr } = useListSources(name, namespace);
  const remove = useRemoveApp();

  const handleRemove = () => {
    remove.mutateAsync({ name: data?.app.name || name, namespace }).then(() => {
      navigate.internal(V2Routes.ApplicationList);
    });
  };

  const kustomizations = kustomizationRes?.kustomizations;

  const errArray = _.compact([error, kustErr, sourcesErr]);
  const isError = errArray.length > 0;

  return (
    <Page
      className={className}
      error={isError ? errArray : null}
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
        <AutomationsTable
          automations={_.map(kustomizations, (k) => ({
            name: k.name,
            type: AutomationKind.Kustomize,
          }))}
          appName={name}
        />
      </Spacer>
      <Spacer m={["small"]}>
        <Flex wide align between>
          <h3>Sources</h3>
          <AddSourceButton appName={name} />
        </Flex>
        <SourcesTable appName={name} sources={sources} />
      </Spacer>
    </Page>
  );
}

export default styled(Application).attrs({ className: Application.name })`
  ${FancyCard} {
    max-width: 272px;
  }
`;
