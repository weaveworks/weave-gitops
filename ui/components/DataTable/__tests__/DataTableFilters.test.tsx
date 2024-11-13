import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import _ from "lodash";
import React from "react";
import { withContext, withTheme } from "../../../lib/test-utils";
import { statusSortHelper } from "../../../lib/utils";
import { filterSeparator } from "../../FilterDialog";
import DataTable from "../DataTable";
import {
  filterByStatusCallback,
  filterConfig,
  filterRows,
  filterSelectionsToQueryString,
  parseFilterStateFromURL,
} from "../helpers";
import { Field } from "../types";

const uriEncodedSeparator = encodeURIComponent(filterSeparator);

const addTextSearchInput = (term: string) => {
  const input = document.getElementById("table-search");
  fireEvent.input(input, { target: { value: term } });
  const form = document.getElementsByTagName("form")[0];
  fireEvent.submit(form);
};

describe("DataTableFilters", () => {
  const rows = [
    {
      name: "cool",
      type: "foo",
      success: false,
      suspended: true,
      conditions: [
        {
          message:
            "Applied revision: main/8868a29b71c008c06549052389f3d762d5fbf821",
          reason: "ReconciliationSucceeded",
          status: "True",
          timestamp: "2022-04-13 20:23:15 +0000 UTC",
          type: "Ready",
        },
      ],
      count: 1,
    },
    {
      name: "slick",
      type: "foo",
      success: true,
      suspended: false,
      conditions: [
        {
          message:
            "Applied revision: main/8868a29b71c008c06549052389f3d762d5fbf821",
          reason: "ReconciliationSucceeded",
          status: "True",
          timestamp: "2022-04-13 20:23:15 +0000 UTC",
          type: "Ready",
        },
      ],
      count: 2,
    },
    {
      name: "neat",
      type: "bar",
      success: false,
      suspended: false,
      conditions: [
        {
          message:
            "Applied revision: main/8868a29b71c008c06549052389f3d762d5fbf821",
          reason: "ArtifactFailed",
          status: "False",
          timestamp: "2022-04-13 20:23:15 +0000 UTC",
          type: "Not Ready",
        },
      ],
      count: 500,
    },
    {
      name: "rad",
      type: "baz",
      success: false,
      suspended: false,
      conditions: [
        {
          message:
            "Applied revision: main/8868a29b71c008c06549052389f3d762d5fbf821",
          reason: "ArtifactFailed",
          status: "False",
          timestamp: "2022-04-13 20:23:15 +0000 UTC",
          type: "Not Ready",
        },
      ],
      count: 500,
    },
  ];

  const fields: Field[] = [
    {
      label: "Name",
      value: "name",
      textSearchable: true,
    },
    {
      label: "Type",
      value: "type",
    },
    {
      label: "Status",
      value: "success",
      sortValue: statusSortHelper,
    },
    {
      label: "Qty",
      value: "count",
    },
  ];
  describe("filterRows", () => {
    it("shows all rows", () => {
      const filtered = filterRows(rows, {});
      expect(filtered).toHaveLength(4);
    });
    it("filters rows", () => {
      const filtered = filterRows(rows, { name: { options: ["cool"] } });
      expect(filtered).toHaveLength(1);
    });
    it("filters rows with more than one value in a filter key", () => {
      const filtered = filterRows(rows, {
        name: { options: ["cool", "slick"] },
      });
      expect(filtered).toHaveLength(2);
    });
    it("ANDs between categories", () => {
      const rows = [
        { name: "a", namespace: "ns1", type: "git" },
        { name: "b", namespace: "ns1", type: "bucket" },
        { name: "c", namespace: "ns2", type: "git" },
      ];
      const filtered = filterRows(rows, {
        namespace: { options: ["ns1"] },
      });
      expect(filtered).toHaveLength(2);

      const filtered2 = filterRows(rows, {
        namespace: { options: ["ns1"] },
        type: { options: ["git"] },
      });
      expect(filtered2).toHaveLength(1);

      const a = _.find(filtered2, { name: "a" });
      const b = _.find(filtered2, { name: "b" });
      const c = _.find(filtered2, { name: "c" });
      expect(a).toBeTruthy();
      expect(b).toBeFalsy();
      expect(c).toBeFalsy();
    });
  });
  it("ANDs between categories, ORs within a category", () => {
    const rows = [
      { name: "a", namespace: "ns1", type: "git" },
      { name: "b", namespace: "ns1", type: "bucket" },
      { name: "c", namespace: "ns2", type: "git" },
    ];
    const filtered = filterRows(rows, {
      namespace: { options: ["ns1", "ns2"] },
    });
    expect(filtered).toHaveLength(3);

    const filtered2 = filterRows(rows, {
      namespace: { options: ["ns1"] },
      type: { options: ["git", "bucket"] },
    });
    expect(filtered2).toHaveLength(2);

    const a2 = _.find(filtered2, { name: "a" });
    const b2 = _.find(filtered2, { name: "b" });
    const c2 = _.find(filtered2, { name: "c" });
    expect(a2).toBeTruthy();
    expect(b2).toBeTruthy();
    expect(c2).toBeFalsy();
  });
  it("should show all rows", () => {
    render(
      withTheme(
        withContext(
          <DataTable fields={fields} rows={rows} filters={{}} />,
          "/applications",
          {}
        )
      )
    );

    expect(screen.queryAllByText("slick")).toBeTruthy();
    expect(screen.queryAllByText("cool")).toBeTruthy();
    expect(screen.queryAllByText("neat")).toBeTruthy();
    expect(screen.queryAllByText("rad")).toBeTruthy();
  });
  it("should filter on click", () => {
    const initialFilterState = {
      ...filterConfig(rows, "type"),
    };

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterState}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );

    const checkbox1 = document.getElementById(
      `type${filterSeparator}foo`
    ) as HTMLInputElement;
    fireEvent.click(checkbox1);

    const tableRows = document.querySelectorAll("tbody tr");

    expect(tableRows).toHaveLength(2);
    expect(tableRows[0].innerHTML).toContain("cool");
    expect(tableRows[1].innerHTML).toContain("slick");

    const chip1 = screen.getByText(`type${filterSeparator}foo`);
    expect(chip1).toBeTruthy();

    const checkbox2 = document.getElementById(
      `type${filterSeparator}baz`
    ) as HTMLInputElement;
    fireEvent.click(checkbox2);

    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(tableRows2).toHaveLength(3);

    expect(tableRows2[2].innerHTML).toContain("rad");

    const chip2 = screen.getByText(`type${filterSeparator}baz`);
    expect(chip2).toBeTruthy();
  });
  it("should filter by status", () => {
    const initialFilterState = {
      ...filterConfig(rows, "status", filterByStatusCallback),
    };

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterState}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );

    const checkbox1 = document.getElementById(
      `status${filterSeparator}Ready`
    ) as HTMLInputElement;
    fireEvent.click(checkbox1);

    const tableRows = document.querySelectorAll("tbody tr");

    expect(tableRows).toHaveLength(1);
    expect(tableRows[0].innerHTML).toContain("slick");

    const chip1 = screen.getByText(`status${filterSeparator}Ready`);
    expect(chip1).toBeTruthy();

    const checkbox2 = document.getElementById(
      `status${filterSeparator}Suspended`
    ) as HTMLInputElement;
    fireEvent.click(checkbox2);

    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(tableRows2).toHaveLength(2);
    expect(tableRows2[0].innerHTML).toContain("cool");

    const chip2 = screen.getByText(`status${filterSeparator}Suspended`);
    expect(chip2).toBeTruthy();
  });
  it("should select/deselect all when category checkbox is clicked", () => {
    const initialFilterState = {
      ...filterConfig(rows, "status", filterByStatusCallback),
    };

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterState}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );

    const checkbox1 = document.getElementById("status") as HTMLInputElement;
    fireEvent.click(checkbox1);
    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(4);
    let checkbox2 = document.getElementById(`status${filterSeparator}Ready`);
    expect(checkbox2).toHaveProperty("checked", true);
    fireEvent.click(checkbox1);
    expect(tableRows).toHaveLength(4);
    checkbox2 = document.getElementById(`status${filterSeparator}Ready`);
    expect(checkbox2).toHaveProperty("checked", false);
  });
  it("should change select all box status when other checkboxes effect state", () => {
    const initialFilterState = {
      ...filterConfig(rows, "status", filterByStatusCallback),
    };

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterState}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );
    const checkbox1 = document.getElementById("status") as HTMLInputElement;
    fireEvent.click(checkbox1);
    let checkbox2 = document.getElementById(`status${filterSeparator}Ready`);
    fireEvent.click(checkbox2);
    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(3);
    expect(checkbox1).toHaveProperty("checked", false);
    checkbox2 = document.getElementById(`status${filterSeparator}Ready`);
    fireEvent.click(checkbox2);
    expect(checkbox1).toHaveProperty("checked", true);
  });
  it("should remove a param when a single chip is clicked", () => {
    const initialFilterState = {
      ...filterConfig(rows, "type"),
    };

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterState}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );

    const checkbox1 = document.getElementById(
      `type${filterSeparator}foo`
    ) as HTMLInputElement;
    fireEvent.click(checkbox1);

    const chip1 = screen.getByText(`type${filterSeparator}foo`);
    expect(chip1).toBeTruthy();
    expect(screen.queryByText("Clear All")).toBeTruthy();

    const tableRows1 = document.querySelectorAll("tbody tr");
    expect(tableRows1).toHaveLength(2);

    // TODO: this is probably an a11y problem. The SVG needs "role=button", since it is clickable
    const svgButton = chip1.parentElement.getElementsByTagName("svg")[0];
    fireEvent.click(svgButton);

    // Should return to all rows being shown
    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(screen.queryByText(`type${filterSeparator}foo`)).toBeFalsy();
    expect(screen.queryByText("Clear All")).toBeFalsy();
    expect(tableRows2).toHaveLength(rows.length);
  });
  it("should clear filtering when the `clear all` chip is clicked", () => {
    const initialFilterState = {
      ...filterConfig(rows, "type"),
    };

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterState}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );

    const checkbox1 = document.getElementById(
      `type${filterSeparator}foo`
    ) as HTMLInputElement;
    fireEvent.click(checkbox1);
    const chip1 = screen.getByText(`type${filterSeparator}foo`);
    expect(chip1).toBeTruthy();

    const tableRows1 = document.querySelectorAll("tbody tr");
    expect(tableRows1).toHaveLength(2);

    const checkbox2 = document.getElementById(
      `type${filterSeparator}baz`
    ) as HTMLInputElement;
    fireEvent.click(checkbox2);
    const chip2 = screen.getByText(`type${filterSeparator}baz`);
    expect(chip2).toBeTruthy();

    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(tableRows2).toHaveLength(3);

    const clearAll = screen.getByText("Clear All");
    // TODO: this is probably an a11y problem. The SVG needs "role=button", since it is clickable
    const svgButton = clearAll.parentElement.getElementsByTagName("svg")[0];
    fireEvent.click(svgButton);

    const tableRows3 = document.querySelectorAll("tbody tr");

    expect(tableRows3).toHaveLength(rows.length);
  });
  it("should add a text filter", () => {
    const initialFilterState = {
      ...filterConfig(rows, "type"),
    };

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterState}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );

    const searchTerms = "my-criterion";
    addTextSearchInput(searchTerms);

    const textChip = screen.queryByText(searchTerms);
    expect(textChip).toBeTruthy();
    expect(textChip.innerHTML).toContain(searchTerms);
  });
  it("should remove a text filter", () => {
    const initialFilterState = {
      ...filterConfig(rows, "type"),
    };

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterState}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );

    const searchTerms = "my-criterion";
    addTextSearchInput(searchTerms);

    const textChip = screen.queryByText(searchTerms);

    const svgButton = textChip.parentElement.getElementsByTagName("svg")[0];
    fireEvent.click(svgButton);

    expect(screen.queryByText(searchTerms)).toBeFalsy();
  });
  it("filters by a text field", () => {
    const initialFilterState = {
      ...filterConfig(rows, "type"),
    };

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterState}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );

    const searchTerms = rows[0].name;
    addTextSearchInput(searchTerms);

    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(1);
    expect(tableRows[0].innerHTML).toContain(searchTerms);
  });
  it("filters by multiple text fields", () => {
    render(
      withTheme(
        withContext(
          <DataTable fields={fields} rows={rows} filters={{}} dialogOpen />,
          "/applications",
          {}
        )
      )
    );

    const term1 = "a";
    addTextSearchInput(term1);

    const term2 = "r";
    addTextSearchInput(term2);

    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(1);
    expect(tableRows[0].innerHTML).toContain(rows[3].name);
  });
  it("filters by fragments of text fields", () => {
    render(
      withTheme(
        withContext(
          <DataTable fields={fields} rows={rows} filters={{}} dialogOpen />,
          "/applications",
          {}
        )
      )
    );

    const row = rows[0];
    const term = row.name.slice(0, 2);
    addTextSearchInput(term);

    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(1);
    expect(tableRows[0].innerHTML).toContain(row.name);
  });
  it("adds filter selection from a URL", () => {
    const initialFilterConfig = {
      ...filterConfig(rows, "type"),
    };

    const search = `?filters=type${uriEncodedSeparator}foo_&search=cool_`;

    render(
      withTheme(
        withContext(
          <DataTable
            fields={fields}
            rows={rows}
            filters={initialFilterConfig}
            dialogOpen
          />,
          "/applications" + search,
          {}
        )
      )
    );
    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(1);
    expect(tableRows[0].innerHTML).toContain("cool");
  });
  it("returns a query string on filter change", () => {
    const queryString = filterSelectionsToQueryString({
      [`type${filterSeparator}foo`]: true,
    });
    expect(queryString).toEqual(`filters=type${uriEncodedSeparator}foo_`);
    expect(parseFilterStateFromURL(queryString)).toEqual({
      initialSelections: { [`type${filterSeparator}foo`]: true },
      textFilters: [],
    });
  });
});
