import { Tabs } from "@mui/material";
import _ from "lodash";
import qs from "query-string";
import * as React from "react";
import { Navigate, Route, Routes, useLocation } from "react-router";
import styled from "styled-components";
import { formatURL } from "../lib/nav";
import Flex from "./Flex";
import Link from "./Link";
import MuiTab from "./MuiTab";

type Props = {
  className?: string;
  children?: any;
  rootPath: string;
  clearQuery?: boolean;
};

type TabProps = {
  name: string;
  path: string;
  children: React.ReactElement<any>;
};

type PathConfig = { name: string; path: string };

const ForwardedLink = React.forwardRef<HTMLAnchorElement>((props, ref) => (
  <Link {...props} ref={ref} />
));

function findChildren(childrenProp: any[]) {
  if (_.isArray(childrenProp)) {
    const childs: any[] = [];
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
  return childrenProp;
}

function routesToIndex(routes: PathConfig[], pathname: string | string[]) {
  const index = _.findIndex(routes, (r) => pathname.includes(r.path));
  return index === -1 ? 0 : index;
}

export function RouterTab({ children }: TabProps) {
  return children;
}

function SubRouterTabs({ className, children, rootPath, clearQuery }: Props) {
  const location = useLocation();
  const query = clearQuery ? "" : qs.parse(location.search);
  const childs = findChildren(children);

  if (!_.get(childs, [0, "props", "path"])) {
    throw new Error("HashRouterTabs children must be of type HashRouterTab");
  }

  const routes: PathConfig[] = _.map(childs, (c: any) => ({
    path: c?.props?.path,
    name: c?.props?.name,
  }));

  return (
    <Flex wide tall column start className={className} gap="12">
      <Tabs
        indicatorColor="primary"
        value={routesToIndex(routes, location.pathname)}
        className="horizontal-tabs"
      >
        {_.map(routes, (route, i) => {
          return (
            <MuiTab
              component={ForwardedLink as typeof Link}
              key={i}
              to={formatURL(`../${route.path}`, query)}
              active={location.pathname.includes(route.path)}
              text={route.name}
            />
          );
        })}
      </Tabs>
      <Routes>
        <Route
          path="/"
          element={<Navigate to={formatURL(rootPath, query)} replace />}
        />
        {_.map(childs, (child, i) => {
          return <Route key={i} path={child.props.path} element={child} />;
        })}
        ;
      </Routes>
    </Flex>
  );
}

export default styled(SubRouterTabs).attrs({ className: SubRouterTabs.name })`
  //MuiTabs
  .horizontal-tabs {
    min-height: ${(props) => props.theme.spacing.large};
    width: 100%;
    .MuiTabs-flexContainer {
      border-bottom: 3px solid ${(props) => props.theme.colors.neutral20};

      .MuiTab-root {
        line-height: 1;
        letter-spacing: 1px;
        min-height: 32px;
        width: fit-content;
        @media (min-width: 600px) {
          min-width: auto;
        }
        @media (min-width: 1440px) {
          min-width: 132px;
        }
        &:hover {
          background-color: ${(props) => props.theme.colors.blueWithOpacity};
          color: ${(props) => props.theme.colors.primary10};
          font-weight: 600;
          span {
            color: ${(props) => props.theme.colors.primary};
          }
        }
      }
    }
    .MuiTabs-indicator {
      height: 0;
      border-block-end: 3px solid ${(props) => props.theme.colors.primary};
    }
  }
`;
