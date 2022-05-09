import _ from "lodash";
import qs from "query-string";
import * as React from "react";
import { useHistory, useLocation } from "react-router-dom";
import styled from "styled-components";
import { IconButton } from "./Button";
import ChipGroup from "./ChipGroup";
import DataTable, { Field } from "./DataTable";
import FilterDialog, {
  DialogFormState,
  FilterConfig,
  formStateToFilters,
  initialFormState,
} from "./FilterDialog";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import { computeReady } from "./KubeStatusIndicator";
import SearchField from "./SearchField";

type Props = {
  className?: string;
  fields: Field[];
  rows: any[];
  filters: FilterConfig;
  dialogOpen?: boolean;
  onDialogClose?: () => void;
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

  return _.filter(rows, (r) => {
    let ok = false;

    _.each(filters, (vals, key) => {
      let value;
      //status
      if (key === "status") {
        if (r["suspended"]) value = "Suspended";
        else if (computeReady(r["conditions"])) value = "Ready";
        else value = "Not Ready";
      }
      //string
      else value = r[key];

      if (_.includes(vals, value)) {
        ok = true;
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

function toPairs(state: State): string[] {
  const result = _.map(state.formState, (val, key) => (val ? key : null));
  const out = _.compact(result);
  return _.concat(out, state.textFilters);
}

type State = {
  filters: FilterConfig;
  formState: DialogFormState;
  textFilters: string[];
};

function FilterableTable({
  className,
  fields,
  rows,
  filters,
  dialogOpen,
}: Props) {
  const history = useHistory();
  const location = useLocation();
  const [filterDialogOpen, setFilterDialogOpen] = React.useState(dialogOpen);
  const [filterState, setFilterState] = React.useState<State>({
    filters,
    formState: initialFormState(filters),
    textFilters: [],
  });
  let filtered = filterRows(rows, filterState.filters);
  filtered = filterText(filtered, fields, filterState.textFilters);
  const chips = toPairs(filterState);

  const handleChipRemove = (chips: string[]) => {
    const next = {
      ...filterState,
    };

    _.each(chips, (chip) => {
      next.formState[chip] = false;
    });

    const filters = formStateToFilters(next.formState);

    const textFilters = _.filter(
      next.textFilters,
      (f) => !_.includes(chips, f)
    );
    setUrl(next.formState);
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
    setUrl(resetFormState);
  };

  const handleFilterSelect = (filters, formState) => {
    setUrl(formState);
    setFilterState({ ...filterState, filters, formState });
  };

  const setUrl = (formState) => {
    let url = "";
    _.each(formState, (value, key) => {
      if (value) {
        url += `${key}_`;
      }
    });
    const query = location.search;
    let prefix = "";
    if (query && !query.includes("filters") && url) prefix = "&?filters=";
    else if (url) prefix = "?filters=";
    history.replace(
      (location.pathname || "") + prefix + encodeURIComponent(url)
    );
  };

  React.useEffect(() => {
    const filterQuery = qs.parse(location.search)["filters"] as string;
    if (filterQuery) {
      const split = filterQuery.split("_");
      const next = filterState.formState;
      _.each(split, (filterString) => {
        if (filterString) next[filterString] = true;
      });
      setFilterState({ ...filterState, formState: next });
    }
  }, []);

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
