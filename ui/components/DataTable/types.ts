import { FilterConfig, FilterSelections } from "../FilterDialog";

export type Field = {
  label: string | number;
  labelRenderer?: string | ((k: any) => string | JSX.Element);
  value: string | ((k: any) => string | JSX.Element | null);
  sortValue?: (k: any) => any;
  textSearchable?: boolean;
  minWidth?: number;
  maxWidth?: number;
  /** boolean for field to initially sort against. */
  defaultSort?: boolean;
  /** boolean for field to implement secondary sort against. */
  secondarySort?: boolean;
};

export type FilterState = {
  filters: FilterConfig;
  formState: FilterSelections;
  textFilters: string[];
};
