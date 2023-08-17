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
import { ThemeTypes } from "../../contexts/AppContext";
import { IconButton } from "../Button";
import CheckboxActions from "../CheckboxActions";
import ChipGroup from "../ChipGroup";
import FilterDialog, {
  FilterConfig,
  FilterSelections,
  selectionsToFilters,
} from "../FilterDialog";
import Flex from "../Flex";
import Icon, { IconType } from "../Icon";
import SearchField from "../SearchField";
import Spacer from "../Spacer";
import Text from "../Text";
import SortableLabel from "./SortableLabel";
import {
  filterRows,
  filterSelectionsToQueryString,
  filterText,
  initialFormState,
  parseFilterStateFromURL,
  sortByField,
  toPairs,
} from "./helpers";
import { Field, FilterState } from "./types";
/** DataTable Properties  */
export interface Props {
  /** The ID of the table. */
  id?: string;
  /** CSS MUI Overrides or other styling. */
  className?: string;
  /** A list of objects with four fields: `label`, which is a string representing the column header, `value`, which can be a string, or a function that extracts the data needed to fill the table cell, and `sortValue`, which customizes your input to the search function */
  fields: Field[];
  /** A list of data that will be iterated through to create the columns described in `fields`. */
  rows?: any[];
  filters?: FilterConfig;
  dialogOpen?: boolean;
  hasCheckboxes?: boolean;
  hideSearchAndFilters?: boolean;
  emptyMessagePlaceholder?: React.ReactNode;
  onColumnHeaderClick?: (field: Field) => void;
  disableSort?: boolean;
}
//styled components
const EmptyRow = styled(TableRow)<{ colSpan: number }>`
  td {
    text-align: center;
  }
`;

const TopBar = styled(Flex)`
  max-width: 100%;
`;

const IconFlex = styled(Flex)`
  position: relative;
  padding: 0 ${(props) => props.theme.spacing.small};
`;

/** Form DataTable */
function UnstyledDataTable({
  id,
  className,
  fields,
  rows,
  filters,
  hasCheckboxes: checkboxes,
  dialogOpen,
  hideSearchAndFilters,
  emptyMessagePlaceholder,
  onColumnHeaderClick,
  disableSort,
}: Props) {
  //URL info
  const history = useHistory();
  const location = useLocation();
  const search = location.search;
  const state = parseFilterStateFromURL(search);

  const [filterDialogOpen, setFilterDialogOpen] = React.useState(dialogOpen);
  const [filterState, setFilterState] = React.useState<FilterState>({
    filters: selectionsToFilters(state.initialSelections, filters),
    formState: initialFormState(filters, state.initialSelections),
    textFilters: state.textFilters,
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

  const handleChipRemove = (chips: string[], filterList) => {
    const next = {
      ...filterState,
    };

    _.each(chips, (chip) => {
      next.formState[chip] = false;
    });

    const filters = selectionsToFilters(next.formState, filterList);

    const textFilters = _.filter(
      next.textFilters,
      (f) => !_.includes(chips, f)
    );

    let query = qs.parse(search);

    if (textFilters.length) query["search"] = textFilters.join("_") + "_";
    else if (query["search"]) query = _.omit(query, "search");
    history.replace({ ...location, search: qs.stringify(query) });

    doChange(next.formState);
    setFilterState({ formState: next.formState, filters, textFilters });
  };

  const handleTextSearchSubmit = (val: string) => {
    if (!val) return;
    const query = qs.parse(search);
    if (!query["search"]) query["search"] = `${val}_`;
    if (!query["search"].includes(val)) query["search"] += `${val}_`;
    history.replace({ ...location, search: qs.stringify(query) });
    setFilterState({
      ...filterState,
      textFilters: _.uniq([...filterState.textFilters, val]),
    });
  };

  const handleClearAll = () => {
    const resetFormState = initialFormState(filters);
    setFilterState({
      filters: {},
      formState: resetFormState,
      textFilters: [],
    });
    const url = qs.parse(location.search);
    //keeps things like clusterName and namespace for details pages
    const cleared = _.omit(url, ["filters", "search"]);
    history.replace({ ...location, search: qs.stringify(cleared) });
  };

  const handleFilterSelect = (filters, formState) => {
    doChange(formState);
    setFilterState({ ...filterState, filters, formState });
  };

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

  const sorted = sortByField(
    filtered,
    reverseSort,
    sortFields,
    useSecondarySort,
    disableSort
  );

  const [checked, setChecked] = React.useState([]);

  const r = _.map(sorted, (r, i) => {
    return (
      <TableRow key={r.uid || i}>
        {checkboxes && (
          <TableCell style={{ padding: "0px" }}>
            <Checkbox
              checked={_.includes(checked, r.uid)}
              onChange={(e) => {
                if (e.target.checked) setChecked([...checked, r.uid]);
                else setChecked(_.without(checked, r.uid));
              }}
              color="primary"
            />
          </TableCell>
        )}
        {_.map(fields, (f) => {
          const style: React.CSSProperties = {
            ...(f.minWidth && { minWidth: f.minWidth }),
            ...(f.maxWidth && { maxWidth: f.maxWidth }),
          };

          return (
            <TableCell
              style={Object.keys(style).length > 0 ? style : undefined}
              key={f.label}
            >
              <Text>
                {(typeof f.value === "function" ? f.value(r) : r[f.value]) ||
                  "-"}
              </Text>
            </TableCell>
          );
        })}
      </TableRow>
    );
  });

  React.useEffect(() => {
    return () => {
      const url = qs.parse(location.search);
      const clearFilters = _.omit(url, ["filters", "search"]);
      history.replace({
        ...history.location,
        search: qs.stringify(clearFilters),
      });
    };
  }, [history]);
  return (
    <Flex wide tall column className={className}>
      <TopBar wide align end>
        {checkboxes && <CheckboxActions checked={checked} rows={filtered} />}
        {filters && !hideSearchAndFilters && (
          <>
            <ChipGroup
              chips={chips}
              onChipRemove={(chips) => handleChipRemove(chips, filters)}
              onClearAll={handleClearAll}
            />
            <IconFlex align>
              <SearchField onSubmit={handleTextSearchSubmit} />
              <IconButton
                onClick={() => setFilterDialogOpen(!filterDialogOpen)}
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
          </>
        )}
      </TopBar>
      <Flex wide tall>
        <TableContainer id={id}>
          <Table aria-label="simple table">
            <TableHead>
              <TableRow>
                {checkboxes && (
                  <TableCell key={"checkboxes"}>
                    <Checkbox
                      checked={filtered.length === checked.length}
                      onChange={(e) =>
                        e.target.checked
                          ? setChecked(filtered.map((r) => r.uid))
                          : setChecked([])
                      }
                      color="primary"
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
                        setSortFieldIndex={(...args) => {
                          if (onColumnHeaderClick) {
                            onColumnHeaderClick(f);
                          }

                          setSortFieldIndex(...args);
                        }}
                        setReverseSort={(isReverse) => {
                          if (onColumnHeaderClick) {
                            onColumnHeaderClick(f);
                          }

                          setReverseSort(isReverse);
                        }}
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
                      {emptyMessagePlaceholder || (
                        <Text color="neutral30">No data</Text>
                      )}
                    </Flex>
                  </TableCell>
                </EmptyRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
        {!hideSearchAndFilters && (
          <FilterDialog
            onFilterSelect={handleFilterSelect}
            filterList={filters}
            formState={filterState.formState}
            open={filterDialogOpen}
          />
        )}
      </Flex>
    </Flex>
  );
}
export const DataTable = styled(UnstyledDataTable)`
  width: 100%;
  flex-wrap: nowrap;
  overflow-x: hidden;
  h2 {
    padding: ${(props) => props.theme.spacing.xs};
    font-size: 12px;
    font-weight: 600;
    color: ${(props) => props.theme.colors.neutral30};
    margin: 0px;
    white-space: nowrap;
    text-transform: uppercase;
    letter-spacing: 1px;
  }
  .MuiTableRow-root {
    transition: background 0.5s ease-in-out;
  }
  .MuiTableRow-root:not(.MuiTableRow-head):hover {
    background: ${(props) =>
      props.theme.mode === ThemeTypes.Dark
        ? props.theme.colors.blueWithOpacity
        : props.theme.colors.neutral10};
    transition: background 0.5s ease-in-out;
  }
  table {
    margin-top: ${(props) => props.theme.spacing.small};
  }
  th {
    padding: 0;
    background: ${(props) => props.theme.colors.neutralGray};
    .MuiCheckbox-root {
      padding: 4px 9px;
    }
    :first-child {
      border-top-left-radius: ${(props) => props.theme.spacing.xxs};
    }
    :last-child {
      border-top-right-radius: ${(props) => props.theme.spacing.xxs};
    }
  }
  td {
    //24px matches th + button + h2 padding
    padding-left: ${(props) => props.theme.spacing.base};
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  //override so filter dialog button stays highlighted, but color is too bright in dark mode
  .MuiButton-contained {
    ${(props) =>
      props.theme.mode === ThemeTypes.Dark
        ? `background-color: ${props.theme.colors.neutral10};`
        : null}
  }
`;

export default DataTable;
