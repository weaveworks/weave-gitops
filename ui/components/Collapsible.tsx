import { Collapse } from "@material-ui/core";
import React from "react";
import styled from "styled-components";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Text from "./Text";

const Collapsible = ({ children }) => {
  const [isOpen, setIsOpen] = React.useState(false);

  const toggle = () => setIsOpen(!isOpen);

  return (
    <div onClick={toggle} style={{ width: "100%" }}>
      <Flex column wide align>
        <div
          style={{
            width: "100%",
            padding: "16px 4px",
            background: "#f6f7f9",
            borderRadius: "4px",
            cursor: "pointer",
          }}
        >
          <Flex wide align gap="16">
            <Icon
              type={
                isOpen
                  ? IconType.KeyboardArrowDownIcon
                  : IconType.KeyboardArrowRightIcon
              }
              size="medium"
              color="neutral40"
            />
            <Text color="neutral30">More Information</Text>
          </Flex>
        </div>
        <Collapse in={isOpen} style={{ width: "100%" }}>
          {children}
        </Collapse>
      </Flex>
    </div>
  );
};

export default styled(Collapsible).attrs({})``;
