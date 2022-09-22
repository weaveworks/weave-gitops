import { Chip } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { filterSeparator } from "./FilterDialog";
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
    <Flex className={className} wide align start>
      {_.map(chips, (chip, index) => {
        const isUndefined =
          //javascript search finds first occurance of substring, returning index
          chip.search(filterSeparator) ===
          //if first occurance of filterSeparator is the end of the chip string, it's an undefined value
          chip.length - filterSeparator.length;
        return (
          <Flex key={index}>
            <Chip
              label={isUndefined ? chip + "null" : chip}
              onDelete={() => onChipRemove([chip])}
            />
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
  height: 40px;
  padding: 4px 0px;
  flex-wrap: nowrap;
  overflow-x: auto;
`;
