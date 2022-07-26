import * as React from "react";
import { useHistory, useLocation } from "react-router-dom";
import styled from "styled-components";
import FilterableTable, {
  FilterableTableProps,
  filterSelectionsToQueryString,
  parseFilterStateFromURL,
} from "./FilterableTable";
import { FilterSelections } from "./FilterDialog";

type Props = FilterableTableProps;

function URLAddressableTable({ initialSelections, ...rest }: Props) {
  const history = useHistory();
  const location = useLocation();
  const search = location.search;
  console.log(search);

  const handleFilterChange = (sel: FilterSelections) => {
    const filterQuery = filterSelectionsToQueryString(sel);
    history.replace({ ...location, search: filterQuery });
  };

  return (
    <FilterableTable
      onFilterChange={handleFilterChange}
      initialSelections={initialSelections || parseFilterStateFromURL(search)}
      {...rest}
    />
  );
}

export default styled(URLAddressableTable).attrs({
  className: URLAddressableTable.name,
})``;
