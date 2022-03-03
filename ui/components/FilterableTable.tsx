import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import DataTable, { Field } from "./DataTable";
import FilterDialog, { FilterConfig } from "./FilterDialog";
import Flex from "./Flex";

type Props = {
  className?: string;
  fields: Field[];
  rows: any[];
  filters: FilterConfig;
  dialogOpen: boolean;
  onDialogClose: () => void;
};

export function filterRows<T>(rows: T[], filters: FilterConfig) {
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

function FilterableTable({
  className,
  fields,
  rows,
  filters,
  dialogOpen,
  onDialogClose,
}: Props) {
  const [filterState, setFilterState] = React.useState<FilterConfig>(filters);
  const filtered = filterRows(rows, filterState);

  return (
    <Flex>
      <DataTable className={className} fields={fields} rows={filtered} />
      <FilterDialog
        onClose={onDialogClose}
        onFilterSelect={setFilterState}
        filterList={filters}
        open={dialogOpen}
      />
    </Flex>
  );
}

export default styled(FilterableTable).attrs({
  className: FilterableTable.name,
})``;
