import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import DataTable, { Field } from "./DataTable";
import { FilterConfig } from "./FilterDialog";

type Props = {
  className?: string;
  fields: Field[];
  rows: any[];
  filters: FilterConfig;
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

function FilterableTable({ className, fields, rows, filters }: Props) {
  const filtered = filterRows(rows, filters);

  return <DataTable className={className} fields={fields} rows={filtered} />;
}

export default styled(FilterableTable).attrs({
  className: FilterableTable.name,
})``;
