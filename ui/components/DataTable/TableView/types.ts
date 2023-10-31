import { Field } from "../types";

export interface SortField extends Field {
  reverseSort: boolean;
}

export interface SortableLabelViewProps {
  field: Field;
  onSortClick?: (field: SortField) => void;
  sortedField?: SortField;
  setSortedField: React.Dispatch<React.SetStateAction<any>>;
}

export interface TableBodyViewProps {
  rows: any[];
  fields: Field[];
  hasCheckboxes?: boolean;
  checkedFields?: string[];
  emptyMessagePlaceholder?: React.ReactNode;
  onCheckChange?: (checked: boolean, id: string) => void;
}

export interface TableHeaderProps {
  fields: Field[];
  defaultSortedField?: SortField;
  hasCheckboxes?: boolean;
  checked?: boolean;
  onSortChange?: (field: SortField) => void;
  onCheckChange?: (checked: boolean) => void;
}

export interface TableViewProps {
  id?: string;
  className?: string;
  fields: Field[];
  defaultSortedField?: SortField;
  rows?: any[];
  onSortChange?: (field: SortField) => void;
  onBatchCheck?: (uids: string[]) => void;
  hasCheckboxes?: boolean;
  checkedFields?: string[];
  emptyMessagePlaceholder?: React.ReactNode;
  disableSort?: boolean;
}
