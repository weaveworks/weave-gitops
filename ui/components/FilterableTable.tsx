import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import ChipGroup from "./ChipGroup";
import DataTable, { Field } from "./DataTable";
import FilterDialog, {
  DialogFormState,
  FilterConfig,
  formStateToFilters,
  initialFormState,
} from "./FilterDialog";
import FilterDialogButton from "./FilterDialogButton";
import Flex from "./Flex";
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

export function filterConfigForType(rows) {
  const typeFilterConfig = _.reduce(
    rows,
    (r, v) => {
      const t = v.type;

      if (!_.includes(r, t)) {
        r.push(t);
      }

      return r;
    },
    []
  );

  return { type: typeFilterConfig };
}

export function filterConfigForStatus(rows) {
  const typeFilterConfig = _.reduce(
    rows,
    (r, v) => {
      let t;
      if (v.suspended) t = "Suspended";
      else if (computeReady(v.status)) t = "Ready";
      else t = "Failed";

      if (!_.includes(r, t)) {
        r.push(t);
      }

      return r;
    },
    []
  );

  return { type: typeFilterConfig };
}

export function filterRows<T>(rows: T[], filters: FilterConfig) {
  if (_.keys(filters).length === 0) {
    return rows;
  }

  return _.filter(rows, (r) => {
    let ok = false;

    _.each(filters, (vals, key) => {
      const value = r[key];

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
  const [filterDialogOpen, setFilterDialog] = React.useState(dialogOpen);
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

    setFilterState({ formState: next.formState, filters, textFilters });
  };

  const handleTextSearchSubmit = (val: string) => {
    setFilterState({
      ...filterState,
      textFilters: _.uniq(_.concat(filterState.textFilters, val)),
    });
  };

  const handleClearAll = () => {
    setFilterState({
      filters: {},
      formState: initialFormState(filters),
      textFilters: [],
    });
  };

  const handleFilterSelect = (filters, formState) => {
    setFilterState({ ...filterState, filters, formState });
  };

  return (
    <Flex className={className} wide tall column>
      <Flex wide>
        <ChipGroup
          chips={chips}
          onChipRemove={handleChipRemove}
          onClearAll={handleClearAll}
        />
        <Flex align wide end>
          <SearchField onSubmit={handleTextSearchSubmit} />
          <FilterDialogButton
            onClick={() => setFilterDialog(!filterDialogOpen)}
          />
        </Flex>
      </Flex>
      <Flex wide tall>
        <DataTable className={className} fields={fields} rows={filtered} />
        <FilterDialog
          onClose={() => setFilterDialog(!filterDialogOpen)}
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
