import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import ChipGroup from "./ChipGroup";
import DataTable, { Field } from "./DataTable";
import FilterDialog, {
  DialogFormState,
  FilterConfig,
  filterSeparator,
  initialFormState,
} from "./FilterDialog";
import Flex from "./Flex";

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

function toPairs(state: DialogFormState): string[] {
  const result = _.map(state, (val, key) => val && key.split(".").join(":"));
  return _.compact(result);
}

type State = {
  filters: FilterConfig;
  formState: DialogFormState;
};

function FilterableTable({
  className,
  fields,
  rows,
  filters,
  dialogOpen,
  onDialogClose,
}: Props) {
  const [filterState, setFilterState] = React.useState<State>({
    filters,
    formState: initialFormState(filters),
  });
  const filtered = filterRows(rows, filterState.filters);

  const handleChipRemove = (chips: string[]) => {
    const next = {
      ...filterState,
    };

    _.each(chips, (chip) => {
      const [k, v] = chip.split(filterSeparator);

      next.filters[k] = _.filter(filterState[k], (d) => d !== v);
      next.formState[chip] = false;
    });

    setFilterState(next);
  };

  return (
    <div className={className}>
      <ChipGroup
        chips={toPairs(filterState.formState)}
        onChipRemove={handleChipRemove}
        onClearAll={() =>
          setFilterState({ filters: {}, formState: initialFormState(filters) })
        }
      />
      <Flex>
        <DataTable className={className} fields={fields} rows={filtered} />
        <FilterDialog
          onClose={onDialogClose}
          onFilterSelect={(filters, formState) =>
            setFilterState({ filters, formState })
          }
          filterList={filters}
          formState={filterState.formState}
          open={dialogOpen}
        />
      </Flex>
    </div>
  );
}

export default styled(FilterableTable).attrs({
  className: FilterableTable.name,
})``;
