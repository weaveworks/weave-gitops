import { Chip } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";

export interface Props {
  className?: string;
  /** currently checked filter options. Part of a `useState` with `setActiveChips` */
  chips: string[];
  /** the setState function for `activeChips` */
  onChipRemove: (chip: string[]) => void;
  onClearAll: () => void;
}

function ChipGroup({ className, chips = [], onChipRemove, onClearAll }: Props) {
  return (
    <Flex className={className} align start>
      {_.map(chips, (chip, index) => {
        return (
          <Flex key={index}>
            <Chip label={chip} onDelete={() => onChipRemove([chip])} />
          </Flex>
        );
      })}
      {chips.length > 0 && <Chip label="Clear All" onDelete={onClearAll} />}
    </Flex>
  );
}

export default styled(ChipGroup).attrs({ className: ChipGroup.name })`
  .MuiChip-root {
    margin-right: ${(props) => props.theme.spacing.xxs};
  }
`;
