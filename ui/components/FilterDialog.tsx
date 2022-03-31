import { List, ListItem, ListItemIcon, Paper, Slide } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import ControlledForm from "./ControlledForm";
import Flex from "./Flex";
import FormCheckbox from "./FormCheckbox";
import Icon, { IconType } from "./Icon";
import Spacer from "./Spacer";
import Text from "./Text";

export type FilterConfig = { [key: string]: string[] };
export type DialogFormState = { [inputName: string]: boolean };

const SlideContainer = styled.div`
  position: relative;
  width: 0px;
`;

const SlideWrapper = styled.div`
  position: absolute;
  right: 0;
  top: 0;
`;

export const filterSeparator = ":";

export function initialFormState(cfg: FilterConfig) {
  return _.reduce(
    cfg,
    (r, vals, k) => {
      _.each(vals, (v) => {
        r[`${k}${filterSeparator}${v}`] = false;
      });

      return r;
    },
    {}
  );
}

/** Filter Bar Properties */
export interface Props {
  className?: string;
  /** the setState function for `activeFilters` */
  onFilterSelect: (val: FilterConfig, state: DialogFormState) => void;
  /** Object containing column headers + corresponding filter options */
  filterList: FilterConfig;
  formState: DialogFormState;
  onClose?: () => void;
  open?: boolean;
}

export function formStateToFilters(values: DialogFormState): FilterConfig {
  const out = {};
  _.each(values, (v, k) => {
    const [key, val] = k.split(filterSeparator);

    if (v) {
      const el = out[key];

      if (el) {
        el.push(val);
      } else {
        out[key] = [val];
      }
    }
  });

  return out;
}

/** Form Filter Bar */
function UnstyledFilterDialog({
  className,
  onFilterSelect,
  filterList,
  formState,
  onClose,
  open,
}: Props) {
  const onFormChange = (name: string, value: any) => {
    if (onFilterSelect) {
      const next = { ...formState, [name]: value };
      onFilterSelect(formStateToFilters(next), next);
    }
  };

  return (
    <SlideContainer>
      <SlideWrapper>
        <Slide direction="left" in={open} mountOnEnter unmountOnExit>
          <Paper elevation={4}>
            <Flex className={className + " filter-bar"} align start>
              <Spacer padding="medium">
                <Flex wide align between>
                  <Text size="extraLarge" color="neutral30">
                    Filters
                  </Text>
                  <Button variant="text" color="inherit" onClick={onClose}>
                    <Icon
                      type={IconType.ClearIcon}
                      size="large"
                      color="neutral30"
                    />
                  </Button>
                </Flex>
                <ControlledForm
                  state={{ values: formState }}
                  onChange={onFormChange}
                >
                  <List>
                    {_.map(filterList, (options: string[], header: string) => {
                      return (
                        <ListItem key={header}>
                          <Flex column>
                            <Text capitalize size="large" color="neutral30">
                              {header}
                            </Text>
                            <List>
                              {_.map(
                                [...options].sort(),
                                (option: string, index: number) => {
                                  return (
                                    <ListItem key={index}>
                                      <ListItemIcon>
                                        <FormCheckbox
                                          label=""
                                          name={`${header}${filterSeparator}${option}`}
                                        />
                                      </ListItemIcon>
                                      <Text color="neutral30">
                                        {_.toString(option)}
                                      </Text>
                                    </ListItem>
                                  );
                                }
                              )}
                            </List>
                          </Flex>
                        </ListItem>
                      );
                    })}
                  </List>
                </ControlledForm>
              </Spacer>
            </Flex>
          </Paper>
        </Slide>
      </SlideWrapper>
    </SlideContainer>
  );
}

export default styled(UnstyledFilterDialog)`
  .MuiPopover-paper {
    min-width: 450px;
    border-left: 2px solid ${(props) => props.theme.colors.neutral30};
    padding-left: ${(props) => props.theme.spacing.medium};
  }
  .MuiListItem-gutters {
    padding-left: 0px;
  }
  .MuiCheckbox-root {
    padding: 0px;
  }
  .MuiCheckbox-colorSecondary {
    &.Mui-checked {
      color: ${(props) => props.theme.colors.primary};
    }
  }
`;
