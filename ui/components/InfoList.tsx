import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";
import { RowHeader } from "./Policies/Utils/HeaderRows";

export type InfoField = [string, any];

const InfoList = styled(
  ({ items }: { className?: string; items: InfoField[] }) => {
    return (
      <Flex column wide gap="8">
        {items.map(([k, v]) => (
          <RowHeader rowkey={k} value={v} key={k} />
        ))}
      </Flex>
    );
  }
)``;

export default InfoList;
