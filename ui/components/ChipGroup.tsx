import { Chip } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";
import Spacer from "./Spacer";

export interface Props {
  className?: string;
  /** currently checked filter options. Part of a `useState` with `setActiveChips` */
  activeChips: string[];
  /** the setState function for `activeChips` */
  setActiveChips: (newChips: string[]) => void;
}

function ChipGroup({ className, activeChips, setActiveChips }: Props) {
  return (
    <Flex className={className} align start>
      {_.map(activeChips, (filter, index) => {
        return (
          <Flex key={index}>
            <Spacer padding="xxs" />
            <Chip
              label={filter}
              onDelete={() =>
                setActiveChips(
                  activeChips.filter((filterCheck) => filterCheck !== filter)
                )
              }
            />
            <Spacer padding="xxs" />
          </Flex>
        );
      })}
      {activeChips.length > 0 && (
        <Chip label="Clear All" onDelete={() => setActiveChips([])} />
      )}
    </Flex>
  );
}

export default styled(ChipGroup).attrs({ className: ChipGroup.name })``;
