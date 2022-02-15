import {
  Checkbox,
  List,
  ListItem,
  ListItemIcon,
  Popover,
} from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Spacer from "./Spacer";
import Text from "./Text";

/** Filter Bar Properties */
export interface Props {
  className?: string;
  /** currently checked filter options. Part of a `useState` with `setActiveFilters` */
  activeFilters: string[];
  /** the setState function for `activeFilters` */
  setActiveFilters: (newFilters: string[]) => void;
  /** Object containing column headers + corresponding filter options */
  filterList: { [header: string]: string[] };
}

/** Form Filter Bar */
function UnstyledFilterBar({
  className,
  activeFilters,
  setActiveFilters,
  filterList,
}: Props) {
  /** why isn't this a ref? It doesn't work. In The MUI Popover docs they do it with setState too... :( https://mui.com/components/popover/ */
  const [anchorEl, setAnchorEl] = React.useState(null);
  const [showFilters, setShowFilters] = React.useState(false);

  const onCheck = (e, option) => {
    e.target.checked
      ? setActiveFilters([...activeFilters, option])
      : setActiveFilters(
          activeFilters.filter((filterCheck) => filterCheck !== option)
        );
  };

  return (
    <Flex className={className} align start>
      <Button
        variant="text"
        color="inherit"
        onClick={(e) => {
          if (!anchorEl) setAnchorEl(e.currentTarget);
          setShowFilters(!showFilters);
        }}
      >
        <Icon type={IconType.FilterIcon} size="medium" color="neutral30" />
      </Button>
      <Popover
        elevation={0}
        open={showFilters}
        anchorEl={anchorEl}
        anchorOrigin={{ vertical: "bottom", horizontal: "left" }}
        onClose={() => setShowFilters(false)}
      >
        <Spacer padding="medium">
          <Flex wide align between>
            <Text size="extraLarge" color="neutral30">
              Filters
            </Text>
            <Button variant="text" color="inherit">
              <Icon type={IconType.ClearIcon} size="large" color="neutral30" />
            </Button>
          </Flex>
          <List>
            {_.map(filterList, (options: string[], header: string) => {
              return (
                <ListItem key={header}>
                  <Flex column>
                    <Text size="large" color="neutral30">
                      {header}
                    </Text>
                    <List>
                      {options.map((option: string, index: number) => {
                        return (
                          <ListItem key={index}>
                            <ListItemIcon>
                              <Checkbox onChange={(e) => onCheck(e, option)} />
                            </ListItemIcon>
                            <Text color="neutral30">{option}</Text>
                          </ListItem>
                        );
                      })}
                    </List>
                  </Flex>
                </ListItem>
              );
            })}
          </List>
        </Spacer>
      </Popover>
    </Flex>
  );
}

export const FilterBar = styled(UnstyledFilterBar)`
  .MuiPopover-paper {
    min-width: 600px;
    border-left: 2px solid ${(props) => props.theme.colors.neutral30};
  }
  .MuiListItem-gutters {
    padding-left: 0px;
  }
  .PrivateSwitchBase-root {
    padding: 9px 9px 9px 0px;
  }
  .MuiIcon-colorSecondary {
    color: ${(props) => props.theme.colors.primary};
  }
`;

export default FilterBar;
