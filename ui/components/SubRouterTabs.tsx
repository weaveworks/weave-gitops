import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import qs from "query-string";
import * as React from "react";
import { useLocation, useHref, Route, Routes } from "react-router-dom";
import styled from "styled-components";
import { formatURL, Redirect } from "../lib/nav";
import Flex from "./Flex";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

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

function TabWrapper({
  route,
  clearQuery,
  active,
}: {
  route: PathConfig;
  clearQuery?: boolean;
  active?: boolean;
}) {
  const { name, path } = route;
  const query = qs.parse(window.location.search);
  // this calculates a link relative to the current path
  const to = useHref(path);
  return (
    <Tab
      component={ForwardedLink as typeof Link}
      to={formatURL(to, clearQuery ? "" : query)}
      className={`${active && "active-tab"}`}
      label={
        <Text
          size="small"
          uppercase
          bold={Boolean(active)}
          semiBold={!active}
          color={active ? "primary10" : "neutral30"}
        >
          {name}
        </Text>
      }
    />
  );
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
      <Tabs indicatorColor="primary" value={activeIndex}>
        {_.map(routes, (route, i) => {
          return (
            <TabWrapper
              key={route.path}
              route={route}
              clearQuery={clearQuery}
              active={i === activeIndex}
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

export default styled(SubRouterTabs).attrs({ className: SubRouterTabs.name })`
  .active-tab {
    background: ${(props) => props.theme.colors.primary}19;
  }
  .MuiTab-root {
    line-height: 1;
    letter-spacing: 1px;
    height: 32px;
    min-height: 32px;
    width: fit-content;
    @media (min-width: 600px) {
      min-width: 132px;
    }
  }
  //trust me there's both tab and tabS
  .MuiTabs-root {
    min-height: 32px;
  }
  .MuiTabs-fixed {
    height: 32px;
  }
  .MuiTabs-indicator {
    height: 3px;
    background-color: ${(props) => props.theme.colors.primary};
  }
`;
