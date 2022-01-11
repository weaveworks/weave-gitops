import _ from "lodash";
import * as React from "react";
import { HashRouter, Route, Switch } from "react-router-dom";
import styled from "styled-components";
import Button from "../../../components/Button";
import DataTable from "../../../components/DataTable";
import FancyCard from "../../../components/FancyCard";
import Flex from "../../../components/Flex";
import Link from "../../../components/Link";
import Page from "../../../components/Page";
import Spacer from "../../../components/Spacer";
import Text from "../../../components/Text";
import { AppContext } from "../../../contexts/AppContext";
import { useGetApplication, useRemoveApp } from "../../../hooks/apps";
import { V2Routes } from "../../../lib/types";
import EmptyApplication from "./EmptyApplication";

type Props = {
  className?: string;
  name: string;
};

function Application({ className, name }: Props) {
  const { navigate } = React.useContext(AppContext);
  const { data, isLoading, error } = useGetApplication(name);
  const remove = useRemoveApp();

  const handleRemove = () => {
    remove.mutateAsync({ name: data?.app.name || name }).then(() => {
      navigate.internal(V2Routes.ApplicationList);
    });
  };

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
        {data?.app.kustomizations.length > 0 ? (
          <HashRouter>
            <Flex column start>
              {_.map(["components", "overlays", "deployments"], (e) => (
                <Spacer key={e} m={["none", "none", "none", "large"]}>
                  <Link to={e}>
                    <Text
                      style={{
                        textDecoration: window.location.hash.includes(e)
                          ? "underline"
                          : "",
                      }}
                      size="large"
                    >
                      {_.capitalize(e)}
                    </Text>
                  </Link>
                </Spacer>
              ))}
            </Flex>
            <Switch>
              <Route
                exact
                path="/components"
                component={() => (
                  <>
                    <div>
                      <h3>Kustomizations</h3>
                    </div>
                    <DataTable
                      sortFields={["name"]}
                      fields={[{ label: "Name", value: "name" }]}
                      rows={data?.app.kustomizations}
                    />
                  </>
                )}
              />
            </Switch>
          </HashRouter>
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
