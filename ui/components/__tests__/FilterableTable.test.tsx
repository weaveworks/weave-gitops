import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import _ from "lodash";
import React from "react";
import { withContext, withTheme } from "../../lib/test-utils";
import { statusSortHelper } from "../../lib/utils";
import { Field, SortType } from "../DataTable";
import FilterableTable, {
  filterByStatusCallback,
  filterByTypeCallback,
  filterConfig,
  filterRows,
  filterSelectionsToQueryString,
  parseFilterStateFromURL,
} from "../FilterableTable";
import { FilterSelections } from "../FilterDialog";

const addTextSearchInput = (term: string) => {
  const input = document.getElementById("table-search");
  fireEvent.input(input, { target: { value: term } });
  const form = document.getElementsByTagName("form")[0];
  fireEvent.submit(form);
};

describe("FilterableTable", () => {
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
      sortType: SortType.number,
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
      const filtered = filterRows(rows, { name: ["cool"] });
      expect(filtered).toHaveLength(1);
    });
    it("filters rows with more than one value in a filter key", () => {
      const filtered = filterRows(rows, { name: ["cool", "slick"] });
      expect(filtered).toHaveLength(2);
    });
    it("ANDs between categories", () => {
      const rows = [
        { name: "a", namespace: "ns1", type: "git" },
        { name: "b", namespace: "ns1", type: "bucket" },
        { name: "c", namespace: "ns2", type: "git" },
      ];
      const filtered = filterRows(rows, {
        namespace: ["ns1"],
      });
      expect(filtered).toHaveLength(2);

      const filtered2 = filterRows(rows, {
        namespace: ["ns1"],
        type: ["git"],
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
      namespace: ["ns1", "ns2"],
    });
    expect(filtered).toHaveLength(3);

    const filtered2 = filterRows(rows, {
      namespace: ["ns1"],
      type: ["git", "bucket"],
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
          <FilterableTable fields={fields} rows={rows} filters={{}} />,
          "/applications",
          {}
        )
      )
    );

    expect(screen.queryAllByText("slick")).toBeTruthy();
    expect(screen.queryAllByText("cool")).toBeTruthy();
  });
  it("should filter rows", () => {
    render(
      withTheme(
        withContext(
          <FilterableTable
            fields={fields}
            rows={rows}
            filters={{ name: ["slick"] }}
            initialSelections={{ "name:slick": true }}
          />,
          "/applications",
          {}
        )
      )
    );

    expect(screen.queryAllByText("slick")[0]).toBeTruthy();
    expect(screen.queryAllByText("cool")[0]).toBeFalsy();
  });
  it("should filter on click", () => {
    const initialFilterState = {
      ...filterConfig(rows, "type"),
    };

    render(
      withTheme(
        withContext(
          <FilterableTable
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

    const checkbox1 = document.getElementById("type:foo") as HTMLInputElement;
    fireEvent.click(checkbox1);

    const tableRows = document.querySelectorAll("tbody tr");

    expect(tableRows).toHaveLength(2);
    expect(tableRows[0].innerHTML).toContain("cool");
    expect(tableRows[1].innerHTML).toContain("slick");

    const chip1 = screen.getByText("type:foo");
    expect(chip1).toBeTruthy();

    const checkbox2 = document.getElementById("type:baz") as HTMLInputElement;
    fireEvent.click(checkbox2);

    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(tableRows2).toHaveLength(3);
    expect(tableRows2[1].innerHTML).toContain("rad");

    const chip2 = screen.getByText("type:baz");
    expect(chip2).toBeTruthy();
  });
  it("should filter by status", () => {
    const initialFilterState = {
      ...filterConfig(rows, "status", filterByStatusCallback),
    };

    render(
      withTheme(
        withContext(
          <FilterableTable
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
      "status:Ready"
    ) as HTMLInputElement;
    fireEvent.click(checkbox1);

    const tableRows = document.querySelectorAll("tbody tr");

    expect(tableRows).toHaveLength(1);
    expect(tableRows[0].innerHTML).toContain("slick");

    const chip1 = screen.getByText("status:Ready");
    expect(chip1).toBeTruthy();

    const checkbox2 = document.getElementById(
      "status:Suspended"
    ) as HTMLInputElement;
    fireEvent.click(checkbox2);

    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(tableRows2).toHaveLength(2);
    expect(tableRows2[0].innerHTML).toContain("cool");

    const chip2 = screen.getByText("status:Suspended");
    expect(chip2).toBeTruthy();
  });
  it("should filter by type with callback", () => {
    const rows = [
      {
        name: "cool",
        groupVersionKind: {
          kind: "foo",
        },
      },
      {
        name: "slick",
        groupVersionKind: {
          kind: "bar",
        },
      },
      {
        name: "neat",
        groupVersionKind: {
          kind: "bar",
        },
      },
      {
        name: "rad",
        groupVersionKind: {
          kind: "bar",
        },
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
        value: (v: any) => v.groupVersionKind.kind,
      },
      {
        label: "Status",
        value: "success",
      },
      {
        label: "Qty",
        value: "count",
      },
    ];

    const initialFilterState = {
      ...filterConfig(rows, "type", filterByTypeCallback),
    };

    render(
      withTheme(
        withContext(
          <FilterableTable
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

    const tableRows1 = document.querySelectorAll("tbody tr");

    expect(tableRows1).toHaveLength(4);
    expect(tableRows1[0].innerHTML).toContain("foo");

    const checkbox1 = document.getElementById("type:foo") as HTMLInputElement;
    fireEvent.click(checkbox1);

    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(tableRows2).toHaveLength(1);
    expect(tableRows2[0].innerHTML).toContain("foo");

    const chip1 = screen.getByText("type:foo");
    expect(chip1).toBeTruthy();

    const clearAll = screen.getByText("Clear All");
    const svgButton = clearAll.parentElement.getElementsByTagName("svg")[0];
    fireEvent.click(svgButton);

    const tableRows3 = document.querySelectorAll("tbody tr");

    expect(tableRows3).toHaveLength(rows.length);

    const checkbox2 = document.getElementById("type:bar") as HTMLInputElement;
    fireEvent.click(checkbox2);

    const tableRows4 = document.querySelectorAll("tbody tr");
    expect(tableRows4).toHaveLength(3);
    expect(tableRows4[0].innerHTML).toContain("bar");
    expect(tableRows4[1].innerHTML).toContain("bar");
    expect(tableRows4[2].innerHTML).toContain("bar");

    const chip2 = screen.getByText("type:bar");
    expect(chip2).toBeTruthy();
  });
  it("should select/deselect all when category checkbox is clicked", () => {
    const initialFilterState = {
      ...filterConfig(rows, "status", filterByStatusCallback),
    };

    render(
      withTheme(
        withContext(
          <FilterableTable
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
    let checkbox2 = document.getElementById("status:Ready");
    expect(checkbox2).toHaveProperty("checked", true);
    fireEvent.click(checkbox1);
    expect(tableRows).toHaveLength(4);
    checkbox2 = document.getElementById("status:Ready");
    expect(checkbox2).toHaveProperty("checked", false);
  });
  it("should change select all box status when other checkboxes effect state", () => {
    const initialFilterState = {
      ...filterConfig(rows, "status", filterByStatusCallback),
    };

    render(
      withTheme(
        withContext(
          <FilterableTable
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
    let checkbox2 = document.getElementById("status:Ready");
    fireEvent.click(checkbox2);
    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(3);
    expect(checkbox1).toHaveProperty("checked", false);
    checkbox2 = document.getElementById("status:Ready");
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
          <FilterableTable
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

    const checkbox1 = document.getElementById("type:foo") as HTMLInputElement;
    fireEvent.click(checkbox1);

    const chip1 = screen.getByText("type:foo");
    expect(chip1).toBeTruthy();
    expect(screen.queryByText("Clear All")).toBeTruthy();

    const tableRows1 = document.querySelectorAll("tbody tr");
    expect(tableRows1).toHaveLength(2);

    // TODO: this is probably an a11y problem. The SVG needs "role=button", since it is clickable
    const svgButton = chip1.parentElement.getElementsByTagName("svg")[0];
    fireEvent.click(svgButton);

    // Should return to all rows being shown
    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(screen.queryByText("type:foo")).toBeFalsy();
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
          <FilterableTable
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

    const checkbox1 = document.getElementById("type:foo") as HTMLInputElement;
    fireEvent.click(checkbox1);
    const chip1 = screen.getByText("type:foo");
    expect(chip1).toBeTruthy();

    const tableRows1 = document.querySelectorAll("tbody tr");
    expect(tableRows1).toHaveLength(2);

    const checkbox2 = document.getElementById("type:baz") as HTMLInputElement;
    fireEvent.click(checkbox2);
    const chip2 = screen.getByText("type:baz");
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
          <FilterableTable
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
          <FilterableTable
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
          <FilterableTable
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
          <FilterableTable
            fields={fields}
            rows={rows}
            filters={{}}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );

    const term1 = rows[0].name;
    addTextSearchInput(term1);

    const term2 = rows[3].name;
    addTextSearchInput(term2);

    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(2);
    expect(tableRows[0].innerHTML).toContain(term1);
    expect(tableRows[1].innerHTML).toContain(term2);
  });
  it("filters by fragments of text fields", () => {
    render(
      withTheme(
        withContext(
          <FilterableTable
            fields={fields}
            rows={rows}
            filters={{}}
            dialogOpen
          />,
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
  it("adds an initial filter selection state", () => {
    const initialFilterConfig = {
      ...filterConfig(rows, "type"),
    };

    render(
      withTheme(
        withContext(
          <FilterableTable
            fields={fields}
            rows={rows}
            initialSelections={{ "type:foo": true }}
            filters={initialFilterConfig}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );
    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(2);
    expect(tableRows[0].innerHTML).toContain("foo");
  });
  it("adds filter selection from a URL", () => {
    const initialFilterConfig = {
      ...filterConfig(rows, "type"),
    };

    const search = `?filters=type%3Afoo_`;

    render(
      withTheme(
        withContext(
          <FilterableTable
            fields={fields}
            rows={rows}
            initialSelections={parseFilterStateFromURL(search)}
            filters={initialFilterConfig}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );
    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(2);
    expect(tableRows[0].innerHTML).toContain("foo");
  });
  it("returns a query string on filter change", () => {
    const initialFilterConfig = {
      ...filterConfig(rows, "type"),
    };

    const recorder = jest.fn();
    const handler = (sel: FilterSelections) => {
      recorder(sel);
    };

    render(
      withTheme(
        withContext(
          <FilterableTable
            onFilterChange={handler}
            fields={fields}
            rows={rows}
            filters={initialFilterConfig}
            dialogOpen
          />,
          "/applications",
          {}
        )
      )
    );
    const checkbox1 = document.getElementById("type:foo") as HTMLInputElement;
    fireEvent.click(checkbox1);
    const args = recorder.mock.calls[0][0];

    expect(args["type:foo"]).toEqual(true);
    const queryString = filterSelectionsToQueryString(args);
    expect(queryString).toEqual("filters=type%3Afoo_");
    expect(parseFilterStateFromURL(queryString)).toEqual({
      "type:foo": true,
    });
  });
});
