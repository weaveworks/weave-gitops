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
  forward: () => void;
  /** func for skip to last page button */
  skipForward: () => void;
  /** func for back one page button */
  back: () => void;
  /** func for skip to start button */
  skipBack: () => void;
  /** onChange func for perPage select */
  perPage: (value) => void;
  /** options for perPage select */
  perPageOptions: number[];
  /** pagination status */
  current: { start: number; pageTotal: number; outOf: number };
}

function unstyledPagination({
  className,
  forward,
  skipForward,
  back,
  skipBack,
  perPage,
  perPageOptions,
  current,
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
              perPage(e.target.value);
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
        {current.start + 1} - {current.start + current.pageTotal} out of{" "}
        {current.outOf}
      </Text>
      <Spacer padding="base" />
      <Flex>
        <Button
          color="inherit"
          variant="text"
          aria-label="skip to first page"
          disabled={current.start === 0}
          onClick={() => skipBack()}
        >
          <Icon type={IconType.SkipPreviousIcon} size="medium" />
        </Button>
        <Button
          color="inherit"
          variant="text"
          aria-label="back one page"
          disabled={current.start === 0}
          onClick={() => back()}
        >
          <Icon type={IconType.NavigateBeforeIcon} size="medium" />
        </Button>
        <Button
          color="inherit"
          variant="text"
          aria-label="forward one page"
          disabled={current.start + current.pageTotal >= current.outOf}
          onClick={() => forward()}
        >
          <Icon type={IconType.NavigateNextIcon} size="medium" />
        </Button>
        <Button
          color="inherit"
          variant="text"
          aria-label="skip to last page"
          disabled={current.start + current.pageTotal >= current.outOf}
          onClick={() => skipForward()}
        >
          <Icon type={IconType.SkipNextIcon} size="medium" />
        </Button>
      </Flex>
    </Flex>
  );
}

export const Pagination = styled(unstyledPagination)``;

export default Pagination;
