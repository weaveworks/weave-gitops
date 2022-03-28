import { Tab, Tabs } from "@material-ui/core";
// eslint-disable-next-line
import { HashHistory } from "history";
import _ from "lodash";
import * as React from "react";
import { HashRouter, Redirect, Route, Switch } from "react-router-dom";
import styled from "styled-components";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

type Props = {
  className?: string;
  children: any;
  defaultPath: string;
  history: HashHistory;
};

type TabProps = {
  name: string;
  path: string;
  children: React.ReactElement;
};

export function HashRouterTab({ children }: TabProps) {
  return (
    <Route exact path={children.props.path}>
      {children}
    </Route>
  );
}

function activeClassName(route, history) {
  return `${history.location.pathname === route && "active-tab"}`;
}

type PathConfig = { name: string; path: string };

function indexToRoute(routes: PathConfig[], index: number) {
  return routes[index];
}

function routesToIndex(routes: PathConfig[], pathname) {
  const index = _.findIndex(routes, (r) => r.path === pathname);
  return index === -1 ? 0 : index;
}

function findChildren(childrenProp) {
  if (_.isArray(childrenProp)) {
    return childrenProp;
  }
  return [childrenProp];
}

function HashRouterTabs({ className, children, defaultPath, history }: Props) {
  const childs = findChildren(children);

  if (!_.get(childs, [0, "props", "path"])) {
    throw new Error("HashRouterTabs children must be of type HashRouterTab");
  }

  const routes: PathConfig[] = _.map(childs, (c: any) => ({
    path: c?.props?.path,
    name: c?.props?.name,
  }));

  return (
    <div className={className}>
      <div>
        <Tabs
          indicatorColor="primary"
          value={routesToIndex(routes, history.location.pathname)}
          onChange={(e, val) => {
            history.push(_.get(indexToRoute(routes, val), "path"));
          }}
        >
          {_.map(routes, (route, i) => (
            <Link key={i} href={`#${route.path}`}>
              <Tab
                className={activeClassName(route.path, history)}
                label={
                  <Text uppercase bold>
                    {route.name}
                  </Text>
                }
              />
            </Link>
          ))}
        </Tabs>
      </div>
      <div>
        <HashRouter>
          <Spacer padding="small" />
          <Switch>
            {children}
            <Redirect exact from="/" to={defaultPath} />
          </Switch>
        </HashRouter>
      </div>
    </div>
  );
}

export default styled(HashRouterTabs).attrs({
  className: HashRouterTabs.name,
})`
  .MuiTabs-root ${Link} .active-tab {
    background: ${(props) => props.theme.colors.primary}19;
  }
`;
