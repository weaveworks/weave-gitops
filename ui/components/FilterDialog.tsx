import { Checkbox, List, ListItem, ListItemIcon } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import ControlledForm from "./ControlledForm";
import Flex from "./Flex";
import FormCheckbox from "./FormCheckbox";
import Text from "./Text";

export type FilterConfig = { [key: string]: string[] };
export type FilterSelections = { [inputName: string]: boolean };

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
  background: rbga(255, 255, 255, 0.75);
  height: 100%;
  width: 100%;
  border-left: 2px solid ${(props) => props.theme.colors.neutral20};
  padding: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.large};
`;

export const filterSeparator = ":";

type FilterSectionProps = {
  header: string;
  options: string[];
  formState: FilterSelections;
  onSectionSelect: (sectionSelectObject) => void;
};

const FilterSection = ({
  header,
  options,
  formState,
  onSectionSelect,
}: FilterSectionProps) => {
  const compoundKeys = options.map((option) => `${header}:${option}`);
  // every on an empty list is true so check that too
  const all =
    compoundKeys.length > 0 && compoundKeys.every((key) => formState[key]);

  const handleChange = () => {
    const optionKeys = _.map(options, (option) => [
      `${header}${filterSeparator}${option}`,
      !all,
    ]);
    onSectionSelect(_.fromPairs(optionKeys));
  };

  return (
    <ListItem>
      <List>
        <ListItem>
          <ListItemIcon>
            <Checkbox
              disabled={!options.length}
              checked={all}
              onChange={handleChange}
              id={header}
            />
          </ListItemIcon>
          <Text capitalize size="small" color="neutral30" semiBold>
            {convertHeaders(header)}
          </Text>
        </ListItem>
        {options.sort().map((option: string, index: number) => (
          <ListItem key={index}>
            <ListItemIcon>
              <FormCheckbox
                label=""
                name={`${header}${filterSeparator}${option}`}
              />
            </ListItemIcon>
            <Text color="neutral40" size="small">
              {_.toString(option) || "-"}
            </Text>
          </ListItem>
        ))}
      </List>
    </ListItem>
  );
};

/** Filter Bar Properties */
export interface Props {
  className?: string;
  /** the setState function for `activeFilters` */
  onFilterSelect: (val: FilterConfig, state: FilterSelections) => void;
  /** Object containing column headers + corresponding filter options */
  filterList: FilterConfig;
  formState: FilterSelections;
  open?: boolean;
}

export function selectionsToFilters(values: FilterSelections): FilterConfig {
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
type sectionSelectObject = { [header: string]: boolean };
/** Form Filter Bar */
function UnstyledFilterDialog({
  className,
  onFilterSelect,
  filterList = {},
  formState = {},
  open,
}: Props) {
  const onSectionSelect = (object: sectionSelectObject) => {
    if (onFilterSelect) {
      const next = { ...formState, ...object };
      onFilterSelect(selectionsToFilters(next), next);
    }
  };
  const onFormChange = (name: string, value: any) => {
    if (onFilterSelect) {
      const next = { ...formState, [name]: value };
      onFilterSelect(selectionsToFilters(next), next);
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
              {Object.entries(filterList)
                .sort()
                .map(([header, options]) => {
                  return (
                    <FilterSection
                      key={header}
                      header={header}
                      options={options}
                      formState={formState}
                      onSectionSelect={onSectionSelect}
                    />
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
