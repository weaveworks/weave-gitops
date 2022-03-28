import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  children: any;
  value: number;
  index: number;
};

// https://v4.mui.com/components/tabs/
function TabPanel(props: Props) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && children}
    </div>
  );
}

export default styled(TabPanel).attrs({ className: TabPanel.name })``;
