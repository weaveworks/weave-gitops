import { List, ListItem, ListItemIcon } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import ControlledForm from "./ControlledForm";
import Flex from "./Flex";
import FormCheckbox from "./FormCheckbox";
import Text from "./Text";

export type FilterConfig = { [key: string]: string[] };
export type DialogFormState = { [inputName: string]: boolean };

const SlideContainer = styled.div`
  position: relative;
  height: 100%;
  width: 0px;
  left: ${(props) => props.theme.spacing.medium};
  transition-property: width, left;
  transition-duration: 0.5s;
  transition-timing-function: ease-in-out;
  &.open {
    left: 0px;
    width: 350px;
  }
`;

const SlideContent = styled.div`
  // this bg color factors in the opacity of the content container while keeping the filters opaque
  background: rgb(250, 250, 250);
  height: 100%;
  width: 100%;
  border-left: 2px solid ${(props) => props.theme.colors.neutral20};
  padding: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.large};
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

const convertHeaders = (header: string) => {
  if (header === "clusterName") return "cluster";
  return header;
};

/** Form Filter Bar */
function UnstyledFilterDialog({
  className,
  onFilterSelect,
  filterList,
  formState,

  open,
}: Props) {
  const onFormChange = (name: string, value: any) => {
    if (onFilterSelect) {
      const next = { ...formState, [name]: value };
      onFilterSelect(formStateToFilters(next), next);
    }
  };

  return (
    <SlideContainer className={`${open ? "open" : ""}`} data-testid="container">
      <SlideContent>
        <Flex className={className} start column>
          <Flex wide align start>
            <Text size="large" color="neutral30">
              Filters
            </Text>
          </Flex>
          <ControlledForm state={{ values: formState }} onChange={onFormChange}>
            <List>
              {_.map(filterList, (options: string[], header: string) => {
                return (
                  <ListItem key={header}>
                    <Flex column>
                      <Text capitalize size="small" color="neutral30">
                        {convertHeaders(header)}
                      </Text>
                      <List>
                        {_.map(options, (option: string, index: number) => {
                          return (
                            <ListItem key={index}>
                              <ListItemIcon>
                                <FormCheckbox
                                  label=""
                                  name={`${header}${filterSeparator}${option}`}
                                />
                              </ListItemIcon>
                              <Text color="neutral30" size="small">
                                {_.toString(option)}
                              </Text>
                            </ListItem>
                          );
                        })}
                      </List>
                    </Flex>
                  </ListItem>
                );
              })}
            </List>
          </ControlledForm>
        </Flex>
      </SlideContent>
    </SlideContainer>
  );
}

export default styled(UnstyledFilterDialog)`
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
