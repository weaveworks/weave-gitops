import { FormControl, MenuItem, Select } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Spacer from "./Spacer";
import Text from "./Text";

export interface Props {
  /** CSS MUI Overrides or other styling. */
  className?: string;
  /** func for forward one page button */
  onForward: () => void;
  /** func for skip to last page button */
  onSkipForward: () => void;
  /** func for back one page button */
  onBack: () => void;
  /** func for skip to start button */
  onSkipBack: () => void;
  /** onChange func for perPage select */
  onSelect: (value) => void;
  /** options for perPage select */
  perPageOptions?: number[];
  /** starting index */
  index: number;
  /** total rows */
  length: number;
  /** all objects */
  totalObjects: number;
}

function unstyledPagination({
  className,
  onForward,
  onSkipForward,
  onBack,
  onSkipBack,
  onSelect,
  perPageOptions = [25, 50, 75, 100],
  index,
  length,
  totalObjects,
}: Props) {
  return (
    <Flex wide align end className={className}>
      <FormControl>
        <Flex align>
          <label htmlFor="pagination">Rows Per Page: </label>
          <Spacer padding="xxs" />
          <Select
            id="pagination"
            variant="outlined"
            defaultValue={perPageOptions[0]}
            onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
              onSelect(e.target.value);
            }}
          >
            {perPageOptions.map((option, index) => {
              return (
                <MenuItem key={index} value={option}>
                  {option}
                </MenuItem>
              );
            })}
          </Select>
        </Flex>
      </FormControl>
      <Spacer padding="base" />
      <Text>
        {index + 1} - {index + length} out of {totalObjects}
      </Text>
      <Spacer padding="base" />
      <Flex>
        <Button
          color="inherit"
          variant="text"
          aria-label="skip to first page"
          disabled={index === 0}
          onClick={() => onSkipBack()}
        >
          <Icon type={IconType.SkipPreviousIcon} size="medium" />
        </Button>
        <Button
          color="inherit"
          variant="text"
          aria-label="back one page"
          disabled={index === 0}
          onClick={() => onBack()}
        >
          <Icon type={IconType.NavigateBeforeIcon} size="medium" />
        </Button>
        <Button
          color="inherit"
          variant="text"
          aria-label="forward one page"
          disabled={index + length >= totalObjects}
          onClick={() => onForward()}
        >
          <Icon type={IconType.NavigateNextIcon} size="medium" />
        </Button>
        <Button
          color="inherit"
          variant="text"
          aria-label="skip to last page"
          disabled={index + length >= totalObjects}
          onClick={() => onSkipForward()}
        >
          <Icon type={IconType.SkipNextIcon} size="medium" />
        </Button>
      </Flex>
    </Flex>
  );
}

export const Pagination = styled(unstyledPagination)``;

export default Pagination;
