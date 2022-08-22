import {
  Checkbox,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from "@material-ui/core";
import _ from "lodash";
import qs from "query-string";
import * as React from "react";
import { useHistory, useLocation } from "react-router-dom";
import styled from "styled-components";
import Button, { IconButton } from "./Button";
import CheckboxActions from "./CheckboxActions";
import ChipGroup from "./ChipGroup";
import FilterDialog, {
  FilterConfig,
  FilterSelections,
  filterSeparator,
  selectionsToFilters,
} from "./FilterDialog";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import { computeReady, ReadyType } from "./KubeStatusIndicator";
import SearchField from "./SearchField";
import Spacer from "./Spacer";
import Text from "./Text";

export type Field = {
  label: string | number;
  labelRenderer?: string | ((k: any) => string | JSX.Element);
  value: string | ((k: any) => string | JSX.Element | null);
  sortValue?: (k: any) => any;
  textSearchable?: boolean;
  maxWidth?: number;
  /** boolean for field to initially sort against. */
  defaultSort?: boolean;
  /** boolean for field to implement secondary sort against. */
  secondarySort?: boolean;
};

type FilterState = {
  filters: FilterConfig;
  formState: FilterSelections;
  textFilters: string[];
};

/** DataTable Properties  */
export interface Props {
  /** CSS MUI Overrides or other styling. */
  className?: string;
  /** A list of objects with four fields: `label`, which is a string representing the column header, `value`, which can be a string, or a function that extracts the data needed to fill the table cell, and `sortValue`, which customizes your input to the search function */
  fields: Field[];
  /** A list of data that will be iterated through to create the columns described in `fields`. */
  rows?: any[];
  filters?: FilterConfig;
  dialogOpen?: boolean;
  checkboxes?: boolean;
}
//styled components
const EmptyRow = styled(TableRow)<{ colSpan: number }>`
  td {
    text-align: center;
  }
`;

const TableButton = styled(Button)`
  &.MuiButton-root {
    margin: 0;
    text-transform: none;
    letter-spacing: 0;
  }
  &.MuiButton-text {
    min-width: 0px;
    .selected {
      color: ${(props) => props.theme.colors.neutral40};
    }
  }
  &.arrow {
    min-width: 0px;
  }
`;

const IconFlex = styled(Flex)`
  position: relative;
  padding: 0 ${(props) => props.theme.spacing.small};
`;
//funcs
export const filterByStatusCallback = (v) => {
  if (v.suspended) return "Suspended";
  else if (computeReady(v["conditions"]) === ReadyType.Reconciling)
    return ReadyType.Reconciling;
  else if (computeReady(v["conditions"]) === ReadyType.Ready)
    return ReadyType.Ready;
  else return ReadyType.NotReady;
};

export const filterByTypeCallback = (v) => _.get(v, "groupVersionKind.kind");

export function filterConfig(
  rows,
  key: string,
  computeValue?: (k: any) => any
): FilterConfig {
  const config = _.reduce(
    rows,
    (r, v) => {
      const t = computeValue ? computeValue(v) : v[key];
      if (!_.includes(r, t)) {
        r.push(t);
      }

      return r;
    },
    []
  );

  return { [key]: config };
}

export function filterRows<T>(rows: T[], filters: FilterConfig) {
  if (_.keys(filters).length === 0) {
    return rows;
  }

  return _.filter(rows, (row) => {
    let ok = true;

    _.each(filters, (vals, category) => {
      let value;
      // status
      if (category === "status") {
        value = filterByStatusCallback(row);
      }
      // type
      else if (category === "type" && typeof row[category] !== "string") {
        value = filterByTypeCallback(row);
      }
      // strings
      else value = row[category];

      if (!_.includes(vals, value)) {
        ok = false;
      }
    });

    return ok;
  });
}

function filterText(
  rows,
  fields: Field[],
  textFilters: FilterState["textFilters"]
) {
  if (textFilters.length === 0) {
    return rows;
  }

  return _.filter(rows, (row) => {
    let matches = false;
    for (const colName in row) {
      const value = row[colName];

      const field = _.find(fields, (f) => {
        if (typeof f.value === "string") {
          return f.value === colName;
        }

        if (f.sortValue) {
          return f.sortValue(row) === value;
        }
      });

      if (!field || !field.textSearchable) {
        continue;
      }

      // This allows us to look for a fragment in the string.
      // For example, if the user searches for "pod", the "podinfo" kustomization should be returned.
      for (const filterString of textFilters) {
        if (_.includes(value, filterString)) {
          matches = true;
        }
      }
    }

    return matches;
  });
}

export function initialFormState(cfg: FilterConfig, initialSelections?) {
  if (!initialSelections) {
    return {};
  }
  const allFilters = _.reduce(
    cfg,
    (r, vals, k) => {
      _.each(vals, (v) => {
        const key = `${k}${filterSeparator}${v}`;
        const selection = _.get(initialSelections, key);
        if (selection) {
          r[key] = selection;
        } else {
          r[key] = false;
        }
      });

      return r;
    },
    {}
  );
  return allFilters;
}

function toPairs(state: FilterState): string[] {
  const result = _.map(state.formState, (val, key) => (val ? key : null));
  const out = _.compact(result);
  return _.concat(out, state.textFilters);
}

export function parseFilterStateFromURL(search: string): FilterSelections {
  const query = qs.parse(search) as any;
  if (query.filters) {
    const split = query.filters.split("_");
    const next = {};
    _.each(split, (filterString) => {
      if (filterString) next[filterString] = true;
    });
    return next;
  }
  return null;
}

export function filterSelectionsToQueryString(sel: FilterSelections) {
  let url = "";
  _.each(sel, (value, key) => {
    if (value) {
      url += `${key}_`;
    }
  });
  //this is an object with all the different queries as keys
  let query = qs.parse(location.search);
  //if there are any filters, reassign/create filter query key
  if (url) query["filters"] = url;
  //if the update leaves no filters, remove the filter query key from the object
  else if (query["filters"]) query = _.omit(query, "filters");
  //this turns a parsed search into a legit query string
  return qs.stringify(query);
}

export const sortByField = (
  rows: any[],
  reverseSort: boolean,
  sortFields: Field[],
  useSecondarySort?: boolean
) => {
  const orderFields = [
    sortFields[0],
    ...(useSecondarySort && sortFields.length > 1 && [sortFields[1]]),
  ];

  return _.orderBy(
    rows,
    sortFields.map((s) => {
      return s.sortValue || s.value;
    }),
    orderFields.map((_, index) => {
      // Always sort secondary sort values in the ascending order.
      const sortOrders =
        reverseSort && (!useSecondarySort || index != 1) ? "desc" : "asc";

      return sortOrders;
    })
  );
};
//components
type labelProps = {
  fields: Field[];
  fieldIndex: number;
  sortFieldIndex: number;
  reverseSort: boolean;
  setSortFieldIndex: (index: number) => void;
  setReverseSort: (b: boolean) => void;
};

function SortableLabel({
  fields,
  fieldIndex,
  sortFieldIndex,
  reverseSort,
  setSortFieldIndex,
  setReverseSort,
}: labelProps) {
  const field = fields[fieldIndex];
  const sort = fields[sortFieldIndex];

  return (
    <Flex align start>
      <TableButton
        color="inherit"
        variant="text"
        onClick={() => {
          setReverseSort(sortFieldIndex === fieldIndex ? !reverseSort : false);
          setSortFieldIndex(fieldIndex);
        }}
      >
        <h2 className={sort.label === field.label ? "selected" : ""}>
          {field.label}
        </h2>
      </TableButton>
      <Spacer padding="xxs" />
      {sort.label === field.label ? (
        <Icon
          type={IconType.ArrowUpwardIcon}
          size="base"
          className={reverseSort ? "upward" : "downward"}
        />
      ) : (
        <div style={{ width: "16px" }} />
      )}
    </Flex>
  );
}

/** Form DataTable */
function UnstyledDataTable({
  className,
  fields,
  rows,
  filters,
  checkboxes,
  dialogOpen,
}: Props) {
  //URL info
  const history = useHistory();
  const location = useLocation();
  const search = location.search;
  const initialSelections = parseFilterStateFromURL(search);

  const [filterDialogOpen, setFilterDialogOpen] = React.useState(dialogOpen);
  const [filterState, setFilterState] = React.useState<FilterState>({
    filters: selectionsToFilters(initialSelections),
    formState: initialFormState(filters, initialSelections),
    textFilters: [],
  });

  const handleFilterChange = (sel: FilterSelections) => {
    const filterQuery = filterSelectionsToQueryString(sel);
    history.replace({ ...location, search: filterQuery });
  };

  let filtered = filterRows(rows, filterState.filters);
  filtered = filterText(filtered, fields, filterState.textFilters);
  const chips = toPairs(filterState);

  const doChange = (formState) => {
    if (handleFilterChange) {
      handleFilterChange(formState);
    }
  };

  const handleChipRemove = (chips: string[]) => {
    const next = {
      ...filterState,
    };

    _.each(chips, (chip) => {
      next.formState[chip] = false;
    });

    const filters = selectionsToFilters(next.formState);

    const textFilters = _.filter(
      next.textFilters,
      (f) => !_.includes(chips, f)
    );

    doChange(next.formState);
    setFilterState({ formState: next.formState, filters, textFilters });
  };

  const handleTextSearchSubmit = (val: string) => {
    setFilterState({
      ...filterState,
      textFilters: _.uniq(_.concat(filterState.textFilters, val)),
    });
  };

  const handleClearAll = () => {
    const resetFormState = initialFormState(filters);
    setFilterState({
      filters: {},
      formState: resetFormState,
      textFilters: [],
    });
    doChange(resetFormState);
  };

  const handleFilterSelect = (filters, formState) => {
    doChange(formState);
    setFilterState({ ...filterState, filters, formState });
  };

  const [checked, setChecked] = React.useState<number[]>([]);

  const [sortFieldIndex, setSortFieldIndex] = React.useState(() => {
    let sortFieldIndex = fields.findIndex((f) => f.defaultSort);

    if (sortFieldIndex === -1) {
      sortFieldIndex = 0;
    }

    return sortFieldIndex;
  });

  const secondarySortFieldIndex = fields.findIndex((f) => f.secondarySort);

  const [reverseSort, setReverseSort] = React.useState(false);
  let sortFields = [fields[sortFieldIndex]];

  const useSecondarySort =
    secondarySortFieldIndex > -1 && sortFieldIndex != secondarySortFieldIndex;

  if (useSecondarySort) {
    sortFields = sortFields.concat(fields[secondarySortFieldIndex]);
    sortFields = sortFields.concat(
      fields.filter(
        (_, index) =>
          index != sortFieldIndex && index != secondarySortFieldIndex
      )
    );
  } else {
    sortFields = sortFields.concat(
      fields.filter((_, index) => index != sortFieldIndex)
    );
  }

  const sorted = sortByField(rows, reverseSort, sortFields, useSecondarySort);

  const r = _.map(sorted, (r, i) => (
    <TableRow key={i}>
      {checkboxes && (
        <TableCell>
          <Checkbox
            checked={checked.includes(r)}
            onChange={(e) => {
              if (e.target.checked) setChecked([...checked, r]);
              else {
                let copy = checked;
                copy.splice(copy.indexOf(r), 1);
                setChecked(copy);
              }
            }}
          />
        </TableCell>
      )}
      {_.map(fields, (f) => (
        <TableCell
          style={
            f.maxWidth && {
              maxWidth: f.maxWidth,
            }
          }
          key={f.label}
        >
          <Text>{typeof f.value === "function" ? f.value(r) : r[f.value]}</Text>
        </TableCell>
      ))}
    </TableRow>
  ));

  return (
    <Flex wide tall column className={className}>
      <Flex wide align between>
        {checkboxes && <CheckboxActions checked={checked} />}
        <Flex wide align end>
          <ChipGroup
            chips={chips}
            onChipRemove={handleChipRemove}
            onClearAll={handleClearAll}
          />
          <IconFlex align>
            <SearchField onSubmit={handleTextSearchSubmit} />
            <IconButton
              onClick={() => setFilterDialogOpen(!filterDialogOpen)}
              className={className}
              variant={filterDialogOpen ? "contained" : "text"}
              color="inherit"
            >
              <Icon
                type={IconType.FilterIcon}
                size="medium"
                color="neutral30"
              />
            </IconButton>
          </IconFlex>
        </Flex>
      </Flex>
      <Flex wide tall>
        <div className={className}>
          <TableContainer>
            <Table aria-label="simple table">
              <TableHead>
                <TableRow>
                  {checkboxes && (
                    <TableCell key={"checkboxes"}>
                      <Checkbox
                        checked={checked.length === rows.length}
                        onChange={(e) =>
                          e.target.checked ? setChecked(rows) : setChecked([])
                        }
                      />
                    </TableCell>
                  )}
                  {_.map(fields, (f, index) => (
                    <TableCell key={f.label}>
                      {typeof f.labelRenderer === "function" ? (
                        f.labelRenderer(r)
                      ) : (
                        <SortableLabel
                          fields={fields}
                          fieldIndex={index}
                          sortFieldIndex={sortFieldIndex}
                          reverseSort={reverseSort}
                          setSortFieldIndex={setSortFieldIndex}
                          setReverseSort={(isReverse) =>
                            setReverseSort(isReverse)
                          }
                        />
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              </TableHead>
              <TableBody>
                {r.length > 0 ? (
                  r
                ) : (
                  <EmptyRow colSpan={fields.length}>
                    <TableCell colSpan={fields.length}>
                      <Flex center align>
                        <Icon
                          color="neutral20"
                          type={IconType.RemoveCircleIcon}
                          size="base"
                        />
                        <Spacer padding="xxs" />
                        <Text color="neutral30">No data</Text>
                      </Flex>
                    </TableCell>
                  </EmptyRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </div>
        <FilterDialog
          onFilterSelect={handleFilterSelect}
          filterList={filterState.filters}
          formState={filterState.formState}
          open={filterDialogOpen}
        />
      </Flex>
    </Flex>
  );
}
export const DataTable = styled(UnstyledDataTable)`
  width: 100%;
  flex-wrap: nowrap;
  overflow-x: auto;
  h2 {
    padding: ${(props) => props.theme.spacing.xs};
    font-size: 14px;
    font-weight: 600;
    color: ${(props) => props.theme.colors.neutral30};
    margin: 0px;
    white-space: nowrap;
  }
  .MuiTableRow-root {
    transition: background 0.5s ease-in-out;
  }
  .MuiTableRow-root:not(.MuiTableRow-head):hover {
    background: ${(props) => props.theme.colors.neutral10};
    transition: background 0.5s ease-in-out;
  }
  table {
    margin-top: ${(props) => props.theme.spacing.small};
  }
  th {
    padding: ${(props) => props.theme.spacing.xs};
    background: ${(props) => props.theme.colors.neutralGray};
  }
  td {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
`;

export default DataTable;
