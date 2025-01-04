//funcs
import _ from "lodash";
import qs from "query-string";
import {
  FilterConfig,
  FilterSelections,
  filterSeparator,
} from "../FilterDialog";
import { computeReady, ReadyType } from "../KubeStatusIndicator";
import { Field, FilterState } from "./types";

export const filterByStatusCallback = (v) => {
  if (v.suspended) return ReadyType.Suspended;
  const ready = computeReady(v["conditions"]);
  if (ready === ReadyType.Reconciling) return ReadyType.Reconciling;
  if (ready === ReadyType.Ready) return ReadyType.Ready;
  return ReadyType.NotReady;
};

export function filterConfig(
  rows,
  key: string,
  computeValue?: (k: any) => any,
): FilterConfig {
  const config = _.reduce(
    rows,
    (r, v) => {
      const t = computeValue ? computeValue(v) : v[key];
      if (!_.includes(r, t)) {
        r.push(t);
      }

      return r;
    },
    [],
  );

  return { [key]: { options: config, transformFunc: computeValue } };
}

export function filterRows<T>(rows: T[], filters: FilterConfig) {
  if (_.keys(filters).length === 0) {
    return rows;
  }

  return _.filter(rows, (row) => {
    let ok = true;

    _.each(filters, (vals, category) => {
      let value;

      if (vals.transformFunc) value = vals.transformFunc(row);
      // strings
      else value = row[category];

      if (!_.includes(vals.options, value)) {
        ok = false;
        return ok;
      }
    });

    return ok;
  });
}

export function filterText(
  rows,
  fields: Field[],
  textFilters: FilterState["textFilters"],
) {
  if (textFilters.length === 0) {
    return rows;
  }

  return _.filter(rows, (row) => {
    let matches = false;

    fields.forEach((field) => {
      if (!field.textSearchable) return matches;

      let value;
      if (field.sortValue) {
        value = field.sortValue(row);
      } else {
        value =
          typeof field.value === "function"
            ? field.value(row)
            : row[field.value];
      }

      for (let i = 0; i < textFilters.length; i++) {
        matches = value.includes(textFilters[i]);
        if (!matches) {
          break;
        }
      }
    });

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
      _.each(vals.options, (v) => {
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
    {},
  );
  return allFilters;
}

export function toPairs(state: FilterState): string[] {
  const result = _.map(state.formState, (val, key) => (val ? key : null));
  const out = _.compact(result);
  return _.concat(out, state.textFilters);
}

export function parseFilterStateFromURL(search: string) {
  const query = qs.parse(search) as any;
  const state = { initialSelections: {}, textFilters: [] };
  if (query.filters) {
    const split = query.filters.split("_");
    const next = {};
    _.each(split, (filterString) => {
      if (filterString) next[filterString] = true;
    });
    state.initialSelections = next;
  }
  if (query.search) {
    state.textFilters = query.search.split("_").filter((item) => item);
  }
  return state;
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

export const sortByField = (
  rows: any[],
  reverseSort: boolean,
  sortFields: Field[],
  useSecondarySort?: boolean,
  disableSort?: boolean,
) => {
  if (disableSort) {
    return rows;
  }

  const orderFields = [sortFields[0]];
  if (useSecondarySort && sortFields.length > 1)
    orderFields.push(sortFields[1]);

  return _.orderBy(
    rows,
    sortFields.map((s) => {
      return s.sortValue || s.value;
    }),
    orderFields.map((_, index) => {
      // Always sort secondary sort values in the ascending order.
      const sortOrders =
        reverseSort && (!useSecondarySort || index != 1) ? "desc" : "asc";

      return sortOrders;
    }),
  );
};
