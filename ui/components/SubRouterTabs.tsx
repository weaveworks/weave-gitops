import { Tabs } from "@material-ui/core";
import _ from "lodash";
import qs from "query-string";
import * as React from "react";
import { useLocation, useHref, Route, Routes } from "react-router-dom";
import styled from "styled-components";
import { formatURL, Redirect } from "../lib/nav";
import Flex from "./Flex";
import Link from "./Link";
import MuiTab from "./MuiTab";
import Spacer from "./Spacer";

type Props = {
  className?: string;
  children?: any;
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

function routesToIndex(routes: PathConfig[], pathname: string): number {
  // FIXME: I can't still find a better way to do this in react-router
  // Maybe after we finish the migration we can look at useMatches(?)
  const index = _.findIndex(routes, (r) => pathname.endsWith(r.path));
  return index === -1 ? 0 : index;
}

export function RouterTab({ children }: TabProps) {
  return children;
}


function SubRouterTabs({ className, children, clearQuery }: Props) {
  const query = qs.parse(window.location.search);
  const childs = findChildren(children).filter((c) => c);

  if (!_.get(childs, [0, "props", "path"])) {
    throw new Error("HashRouterTabs children must be of type HashRouterTab");
  }

  const routes: PathConfig[] = _.map(childs, (c: any) => ({
    path: c?.props?.path,
    name: c?.props?.name,
  }));

  const defaultTabPath = useHref(routes[0].path);
  const { pathname } = useLocation();
  const activeIndex = routesToIndex(routes, pathname);

  return (
    <Flex wide tall column start className={className}>
      <Tabs
        indicatorColor="primary"
        value={activeIndex}
        className="horizontal-tabs"
      >
        {_.map(routes, (route, i) => {
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
      <Routes>
        {_.map(childs, (c) => (
          <Route key={c.props.path} path={c.props.path} element={c} />
        ))}
        <Route
          path="*"
          element={
            <Redirect to={formatURL(defaultTabPath, clearQuery ? "" : query)} />
          }
        />
      </Routes>
    </Flex>
  );
}

export default styled(SubRouterTabs).attrs({ className: SubRouterTabs.name })``;
