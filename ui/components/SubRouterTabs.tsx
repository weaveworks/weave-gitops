import { Tabs } from "@material-ui/core";
import _ from "lodash";
import qs from "query-string";
import * as React from "react";
import { Redirect, Route, Switch } from "react-router-dom";
import styled from "styled-components";
import { formatURL } from "../lib/nav";
import Flex from "./Flex";
import Link from "./Link";
import MuiTab from "./MuiTab";
import Spacer from "./Spacer";

type Props = {
  className?: string;
  children?: any;
  rootPath: string;
  clearQuery?: boolean;
};

type TabProps = {
  name: string;
  path: string;
  children: React.ReactElement;
};

type PathConfig = { name: string; path: string };

const ForwardedLink = React.forwardRef((props, ref) => (
  <Link {...props} innerRef={ref} />
));

function findChildren(childrenProp) {
  if (_.isArray(childrenProp)) {
    const childs = [];
    childrenProp.forEach((child) => {
      if (_.isArray(child)) {
        child.forEach((ch) => {
          childs.push(ch);
        });
      } else {
        childs.push(child);
      }
    });
    return childs;
  }
  return [childrenProp];
}

function routesToIndex(routes: PathConfig[], pathname) {
  const index = _.findIndex(routes, (r) => pathname.includes(r.path));
  return index === -1 ? 0 : index;
}

export function RouterTab({ children }: TabProps) {
  return (
    <Route exact path={children.props.path}>
      {children}
    </Route>
  );
}

function SubRouterTabs({ className, children, rootPath, clearQuery }: Props) {
  const query = qs.parse(window.location.search);
  const childs = findChildren(children);

  if (!_.get(childs, [0, "props", "path"])) {
    throw new Error("HashRouterTabs children must be of type HashRouterTab");
  }

  const routes: PathConfig[] = _.map(childs, (c: any) => ({
    path: c?.props?.path,
    name: c?.props?.name,
  }));

  return (
    <Flex wide tall column start className={className}>
      <Tabs
        indicatorColor="primary"
        value={routesToIndex(routes, window.location.pathname)}
        className="horizontal-tabs"
      >
        {_.map(routes, (route, i) => {
          const bold = window.location.pathname.includes(route.path);

          return (
            <MuiTab
              component={ForwardedLink as typeof Link}
              key={i}
              to={formatURL(`${route.path}`, clearQuery ? "" : query)}
              active={window.location.pathname.includes(route.path)}
              text={route.name}
            />
          );
        })}
      </Tabs>
      <Spacer padding="xs" />
      <Switch>
        {children}
        <Redirect from="*" to={formatURL(rootPath, clearQuery ? "" : query)} />
      </Switch>
    </Flex>
  );
}

export default styled(SubRouterTabs).attrs({ className: SubRouterTabs.name })``;
