import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { PageRoute } from "../lib/types";
import { formatURL, getNavValue } from "../lib/utils";
import Link from "./Link";

type Props = {
  className?: string;
};

const navItems = [{ value: PageRoute.Applications, label: "Applications" }];

const LinkTab = (props) => (
  <Tab
    component={React.forwardRef((p: any, ref) => (
      <Link innerRef={ref} {...p} />
    ))}
    {...props}
  />
);

const StyleLinkTab = styled(LinkTab)`
  span {
    align-items: flex-start;
    color: #4b4b4b;
  }
`;

const NavContent = styled.div`
  background-color: white;
  height: 100vh;
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.xs};
  padding-right: ${(props) => props.theme.spacing.xl};

  .MuiTab-wrapper {
    text-transform: capitalize;
    font-size: 20px;
    font-weight: bold;
  }

  .MuiTabs-indicator {
    left: 0;
    width: 4px;
    background-color: ${(props) => props.theme.colors.primary};
  }
`;

function LeftNav({ className }: Props) {
  const { currentPage } = useNavigation();
  return (
    <div className={className}>
      <NavContent>
        <Tabs
          centered={false}
          orientation="vertical"
          value={getNavValue(currentPage)}
        >
          {_.map(navItems, (n) => (
            <StyleLinkTab
              value={n.value}
              key={n.value}
              label={n.label}
              to={formatURL(n.value)}
            />
          ))}
        </Tabs>
      </NavContent>
    </div>
  );
}

export default styled(LeftNav)`
  width: 240px;
  background-color: ${(props) => props.theme.colors.white};
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-right: ${(props) => props.theme.spacing.medium};
  background-color: ${(props) => props.theme.colors.negativeSpace};
`;
