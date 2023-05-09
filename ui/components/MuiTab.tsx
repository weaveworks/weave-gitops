import { Tab } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Text from "./Text";

type Props = {
  className?: string;
  active?: boolean;
  component?: any;
  key?: number;
  to?: string;
  text: string;
  onClick?: any;
};

function MuiTab({
  className,
  active,
  component,
  key,
  to,
  text,
  onClick,
}: Props) {
  return (
    <Tab
      className={className + `${active && " active-tab"}`}
      component={component}
      key={key}
      to={to}
      onClick={onClick}
      label={
        <Text
          size="small"
          uppercase
          bold={active}
          semiBold={active}
          color={active ? "primary10" : "neutral30"}
        >
          {text}
        </Text>
      }
    />
  );
}

export default styled(MuiTab).attrs({ className: MuiTab.name })`
  &.active-tab {
    background: ${(props) => props.theme.colors.primary}19;
  }
`;
