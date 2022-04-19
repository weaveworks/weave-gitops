import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import _ from "lodash";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import { Field } from "../DataTable";
import FilterableTable, {
  filterConfigForStatus,
  filterConfigForString,
  filterRows,
} from "../FilterableTable";

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
          type: "Ready",
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
          type: "Ready",
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
    it("filters rows with multiple filter keys", () => {
      const filtered = filterRows(rows, { name: ["cool"], type: ["bar"] });
      expect(filtered).toHaveLength(2);
      const cool = _.find(filtered, { name: "cool" });
      const neat = _.find(filtered, { name: "neat" });
      const slick = _.find(filtered, { name: "slick" });

      expect(cool).toBeTruthy();
      expect(neat).toBeTruthy();
      expect(slick).toBeFalsy();
    });
    it("filters rows with multiple filters in multiple keys", () => {
      const filtered = filterRows(rows, {
        name: ["cool"],
        type: ["bar", "foo"],
      });
      expect(filtered).toHaveLength(3);

      const filtered2 = filterRows(rows, {
        name: [],
        type: ["baz", "foo"],
      });
      expect(filtered2).toHaveLength(3);

      const neat = _.find(filtered2, { name: "neat" });
      expect(neat).toBeFalsy();
    });
  });

  it("should show all rows", () => {
    render(
      withTheme(<FilterableTable fields={fields} rows={rows} filters={{}} />)
    );

    expect(screen.queryAllByText("slick")).toBeTruthy();
    expect(screen.queryAllByText("cool")).toBeTruthy();
  });
  it("should filter rows", () => {
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={{ name: ["slick"] }}
        />
      )
    );

    expect(screen.queryAllByText("slick")[0]).toBeTruthy();
    expect(screen.queryAllByText("cool")[0]).toBeFalsy();
  });
  it("should filter on click", () => {
    const initialFilterState = {
      ...filterConfigForString(rows, "type"),
    };
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
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
      ...filterConfigForStatus(rows),
    };
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
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
  it("should remove a param when a single chip is clicked", () => {
    const initialFilterState = {
      ...filterConfigForString(rows, "type"),
    };
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
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
    expect(tableRows2).toHaveLength(rows.length);

    expect(screen.queryByText("type:foo")).toBeFalsy();
    expect(screen.queryByText("Clear All")).toBeFalsy();
  });
  it("should clear filtering when the `clear all` chip is clicked", () => {
    const initialFilterState = {
      ...filterConfigForString(rows, "type"),
    };
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
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
      ...filterConfigForString(rows, "type"),
    };
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
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
      ...filterConfigForString(rows, "type"),
    };
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
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
      ...filterConfigForString(rows, "type"),
    };

    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
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
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={{
            ...filterConfigForString(rows, "type"),
          }}
          dialogOpen
        />
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
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={{
            ...filterConfigForString(rows, "type"),
          }}
          dialogOpen
        />
      )
    );

    const row = rows[0];
    const term = row.name.slice(0, 2);
    addTextSearchInput(term);

    const tableRows = document.querySelectorAll("tbody tr");
    expect(tableRows).toHaveLength(1);
    expect(tableRows[0].innerHTML).toContain(row.name);
  });
});
