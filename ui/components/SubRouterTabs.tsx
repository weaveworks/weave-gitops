import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import qs from "query-string";
import * as React from "react";
import {
  useLocation,
  useHref,
  Route,
  Routes,
  useNavigate,
} from "react-router-dom-v5-compat";
import styled from "styled-components";
import { formatURL } from "../lib/nav";
import Flex from "./Flex";
import { routeTab } from "./KustomizationDetail";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

function Redirect({ to }) {
  const navigate = useNavigate();
  React.useEffect(() => {
    navigate(to);
  });
  return null;
}

type Props = {
  className?: string;
  tabs: routeTab[];
  clearQuery?: boolean;
};

const ForwardedLink = React.forwardRef((props, ref) => (
  <Link {...props} innerRef={ref} />
));

export function RouterTab({
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

type PathConfig = { name: string; path: string };

function routesToIndex(routes: PathConfig[], pathname: string): number {
  // FIXME: I can't still find a better way to do this in react-router
  const index = _.findIndex(routes, (r) => pathname.includes(r.path));
  return index === -1 ? 0 : index;
}

function SubRouterTabs({ className, tabs, clearQuery }: Props) {
  const { pathname } = useLocation();
  const defaultTabPath = useHref(tabs[0].path);
  const query = qs.parse(window.location.search);
  const activeIndex = routesToIndex(tabs, pathname);

  return (
    <Flex wide tall column start className={className}>
      <Tabs indicatorColor="primary" value={activeIndex}>
        {_.map(tabs, (route, i) => {
          return (
            <RouterTab
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
        {_.map(tabs, (route) => {
          return (
            <Route
              key={route.path}
              path={route.path}
              element={route.component()}
            />
          );
        })}
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
