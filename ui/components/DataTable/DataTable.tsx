import _ from "lodash";
import qs from "query-string";
import * as React from "react";
import { useNavigate, useLocation } from "react-router";
import styled from "styled-components";
import { ThemeTypes } from "../../contexts/AppContext";
import { SearchedNamespaces } from "../../lib/types";
import { IconButton } from "../Button";
import ChipGroup from "../ChipGroup";
import FilterDialog, {
  FilterConfig,
  FilterSelections,
  selectionsToFilters,
} from "../FilterDialog";
import Flex from "../Flex";
import Icon, { IconType } from "../Icon";
import SearchField from "../SearchField";
import CheckboxActions from "../Sync/CheckboxActions";
import {
  filterRows,
  filterSelectionsToQueryString,
  filterText,
  initialFormState,
  parseFilterStateFromURL,
  toPairs,
} from "./helpers";

import SearchedNamespacesModal from "./TableView/SearchedNamespacesModal";
import TableView from "./TableView/TableView";
import { SortField } from "./TableView/types";
import { Field, FilterState } from "./types";

export interface Props {
  id?: string;
  className?: string;
  fields: Field[];
  rows?: any[];
  filters?: FilterConfig;
  dialogOpen?: boolean;
  hasCheckboxes?: boolean;
  hideSearchAndFilters?: boolean;
  emptyMessagePlaceholder?: React.ReactNode;
  onColumnHeaderClick?: (field: Field) => void;
  disableSort?: boolean;
  searchedNamespaces?: SearchedNamespaces;
}

const TopBar = styled(Flex)`
  max-width: 100%;
`;

const IconFlex = styled(Flex)`
  position: relative;
  padding: 0 ${(props) => props.theme.spacing.small};
`;

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
  searchedNamespaces,
}: Props) {
  const navigate = useNavigate();
  const location = useLocation();
  const search = location.search;
  const state = parseFilterStateFromURL(search);
  const [filterDialogOpen, setFilterDialogOpen] = React.useState(dialogOpen);

  const [checked, setChecked] = React.useState([]);

  const [filterState, setFilterState] = React.useState<FilterState>({
    filters: selectionsToFilters(state.initialSelections, filters),
    formState: initialFormState(filters, state.initialSelections),
    textFilters: state.textFilters,
  });

  const [sortedItem, setSortedItem] = React.useState<SortField | null>(() => {
    const defaultSortField = fields.find((f) => f.defaultSort);
    const sortField = defaultSortField
      ? {
          ...defaultSortField,
          reverseSort: false,
        }
      : null;
    return sortField;
  });

  const handleFilterChange = (sel: FilterSelections) => {
    const filterQuery = filterSelectionsToQueryString(sel);
    navigate({ ...location, search: filterQuery }, { replace: true });
  };

  let filtered = filterRows(rows, filterState.filters);
  filtered = filterText(filtered, fields, filterState.textFilters);
  const chips = toPairs(filterState);

  const sortItems = (filtered) => {
    let sorted = filtered;
    if (sortedItem) {
      sorted = _.orderBy(
        filtered,
        [sortedItem.sortValue || sortedItem.value],
        [sortedItem.reverseSort ? "desc" : "asc"],
      );
    }
    return sorted;
  };

  const items = sortItems(filtered);

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
      (f) => !_.includes(chips, f),
    );

    let query = qs.parse(search);

    if (textFilters.length) query["search"] = textFilters.join("_") + "_";
    else if (query["search"]) query = _.omit(query, "search");
    navigate({ ...location, search: qs.stringify(query) }, { replace: true });

    doChange(next.formState);
    setFilterState({ formState: next.formState, filters, textFilters });
  };

  const handleTextSearchSubmit = (val: string) => {
    if (!val) return;
    const query = qs.parse(search);
    if (!query["search"]) query["search"] = `${val}_`;
    if (!query["search"].includes(val)) query["search"] += `${val}_`;
    navigate({ ...location, search: qs.stringify(query) }, { replace: true });
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
    navigate({ ...location, search: qs.stringify(cleared) }, { replace: true });
  };

  const handleFilterSelect = (filters, formState) => {
    doChange(formState);
    setFilterState({ ...filterState, filters, formState });
  };

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
              {searchedNamespaces && (
                <SearchedNamespacesModal
                  searchedNamespaces={searchedNamespaces}
                />
              )}
              <SearchField onSubmit={handleTextSearchSubmit} />
              <IconButton
                onClick={() => setFilterDialogOpen(!filterDialogOpen)}
                variant={filterDialogOpen ? "contained" : "text"}
                color="inherit"
                size="large"
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
        <TableView
          fields={fields}
          rows={items}
          defaultSortedField={sortedItem}
          id={id}
          hasCheckboxes={checkboxes}
          emptyMessagePlaceholder={emptyMessagePlaceholder}
          checkedFields={checked}
          disableSort={disableSort}
          onBatchCheck={(checked) => setChecked(checked)}
          onSortChange={(field) => {
            if (onColumnHeaderClick) onColumnHeaderClick(field);
            setSortedItem(field);
          }}
        />
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

  .MuiTableCell-root {
    border-color: ${(props) =>
      props.theme.mode === ThemeTypes.Dark
        ? props.theme.colors.primary30
        : props.theme.colors.neutral20};
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
