import {
  Checkbox,
  Chip,
  Divider,
  List,
  ListItem,
  ListItemIcon,
  Popover,
} from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { theme } from "..";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Input from "./Input";
import Spacer from "./Spacer";
import Text from "./Text";

/** Filter Bar Properties */
export interface Props {
  className?: string;
  /** currently checked filter options mapped to chips next to filter icon. Part of a `useState` with `setActiveFilters` */
  activeFilters: string[];
  /** the setState function for `activeFilters` */
  setActiveFilters: (newFilters: string[]) => void;
  /** Object containing column headers + corresponding filter options */
  filterList: Record<string, string[]>;
}

const FilterHeader = styled(Flex)`
  background: ${theme.colors.primary10};
  color: white;
  padding: ${theme.spacing.base};
`;

/** Form Filter Bar */
function UnstyledFilterBar({
  className,
  activeFilters,
  setActiveFilters,
  filterList,
}: Props) {
  const [anchorEl, setAnchorEl] = React.useState(null);
  const [showFilters, setShowFilters] = React.useState(false);
  const [search, setSearch] = React.useState("");

  const onCheck = (e, option) => {
    e.target.checked
      ? setActiveFilters([...activeFilters, option])
      : setActiveFilters(
          activeFilters.filter((filterCheck) => filterCheck !== option)
        );
  };

  return (
    <Flex className={className} align wide start>
      <Button
        variant="text"
        onClick={(e) => {
          if (!anchorEl) setAnchorEl(e.currentTarget);
          setShowFilters(!showFilters);
        }}
      >
        <Icon type={IconType.FilterIcon} size="medium" color="neutral30" />
      </Button>
      <Popover
        className="filter-popover"
        open={showFilters}
        anchorEl={anchorEl}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
        onClose={() => setShowFilters(false)}
      >
        <FilterHeader align start>
          <Text size="large">Filters</Text>
        </FilterHeader>
        <Spacer padding="base">
          <Flex align start>
            <Icon type={IconType.SearchIcon} size="medium" />
            <Spacer padding="xxs" />
            <Input
              placeholder="SEARCH"
              defaultValue={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </Flex>
        </Spacer>
        <Divider />
        <List>
          {Object.keys(filterList).map((header: string, index: number) => {
            return (
              <div key={index}>
                <ListItem>
                  <Flex column>
                    <Text size="large">{header}</Text>
                    <List>
                      {filterList[header].map(
                        (option: string, index: number) => {
                          if (search) {
                            for (let i = 0; i < search.length; i++) {
                              if (
                                search[i].toLowerCase() !==
                                option[i].toLowerCase()
                              ) {
                                return;
                              }
                            }
                          }
                          return (
                            <ListItem key={index}>
                              <ListItemIcon>
                                <Checkbox
                                  onChange={(e) => onCheck(e, option)}
                                />
                              </ListItemIcon>
                              <Text>{option}</Text>
                            </ListItem>
                          );
                        }
                      )}
                    </List>
                  </Flex>
                </ListItem>
                {index < Object.keys(filterList).length - 1 && <Divider />}
              </div>
            );
          })}
        </List>
      </Popover>
      {_.map(activeFilters, (filter, index) => {
        return (
          <Flex key={index}>
            <Chip
              label={filter}
              onDelete={() =>
                setActiveFilters(
                  activeFilters.filter((filterCheck) => filterCheck !== filter)
                )
              }
            />
            <Spacer padding="xxs" />
          </Flex>
        );
      })}
      <Chip label="Clear All" onDelete={() => setActiveFilters([])} />
    </Flex>
  );
}

const FilterBar = styled(UnstyledFilterBar).attrs({
  className: UnstyledFilterBar.name,
})``;
export default FilterBar;
