import _ from "lodash";
import qs from "query-string";
import * as React from "react";
import styled from "styled-components";
import { IconButton } from "./Button";
import ChipGroup from "./ChipGroup";
import DataTable, { Field } from "./DataTable";
import FilterDialog, {
  FilterConfig,
  FilterSelections,
  filterSeparator,
  selectionsToFilters,
} from "./FilterDialog";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import { computeReady } from "./KubeStatusIndicator";
import SearchField from "./SearchField";

export type FilterableTableProps = {
  className?: string;
  fields: Field[];
  rows: any[];
  filters: FilterConfig;
  dialogOpen?: boolean;
  onDialogClose?: () => void;
  initialSelections?: FilterSelections;
  onFilterChange?: (sel: FilterSelections) => void;
};

export function filterConfigForString(rows, key: string) {
  const typeFilterConfig = _.reduce(
    rows,
    (r, v) => {
      const t = v[key];

      if (!_.includes(r, t)) {
        r.push(t);
      }

      return r;
    },
    []
  );

  return { [key]: typeFilterConfig };
}

export function filterConfigForStatus(rows) {
  const statusFilterConfig = _.reduce(
    rows,
    (r, v) => {
      let t;
      if (v.suspended) t = "Suspended";
      else if (computeReady(v.conditions)) t = "Ready";
      else t = "Not Ready";
      if (!_.includes(r, t)) {
        r.push(t);
      }
      return r;
    },
    []
  );

  return { status: statusFilterConfig };
}

export function filterRows<T>(rows: T[], filters: FilterConfig) {
  if (_.keys(filters).length === 0) {
    return rows;
  }

  return _.filter(rows, (row) => {
    let ok = true;

    _.each(filters, (vals, category) => {
      let value;
      //status
      if (category === "status") {
        if (row["suspended"]) value = "Suspended";
        else if (computeReady(row["conditions"])) value = "Ready";
        else value = "Not Ready";
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

function filterText(rows, fields: Field[], textFilters: State["textFilters"]) {
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

function toPairs(state: State): string[] {
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

type State = {
  filters: FilterConfig;
  formState: FilterSelections;
  textFilters: string[];
};

function FilterableTable({
  className,
  fields,
  rows,
  filters,
  dialogOpen,
  initialSelections,
  onFilterChange,
}: FilterableTableProps) {
  const [filterDialogOpen, setFilterDialogOpen] = React.useState(dialogOpen);
  const [filterState, setFilterState] = React.useState<State>({
    filters: selectionsToFilters(initialSelections),
    formState: initialFormState(filters, initialSelections),
    textFilters: [],
  });

  let filtered = filterRows(rows, filterState.filters);
  filtered = filterText(filtered, fields, filterState.textFilters);
  const chips = toPairs(filterState);

  const doChange = (formState) => {
    if (onFilterChange) {
      onFilterChange(formState);
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

  return (
    <Flex className={className} wide tall column>
      <Flex wide align>
        <ChipGroup
          chips={chips}
          onChipRemove={handleChipRemove}
          onClearAll={handleClearAll}
        />
        <Flex align wide end>
          <SearchField onSubmit={handleTextSearchSubmit} />
          <IconButton
            onClick={() => setFilterDialogOpen(!filterDialogOpen)}
            className={className}
            variant={filterDialogOpen ? "contained" : "text"}
            color="inherit"
          >
            <Icon type={IconType.FilterIcon} size="medium" color="neutral30" />
          </IconButton>
        </Flex>
      </Flex>
      <Flex wide tall>
        <DataTable className={className} fields={fields} rows={filtered} />
        <FilterDialog
          onFilterSelect={handleFilterSelect}
          filterList={filters}
          formState={filterState.formState}
          open={filterDialogOpen}
        />
      </Flex>
    </Flex>
  );
}

export default styled(FilterableTable).attrs({
  className: FilterableTable.name,
})``;
