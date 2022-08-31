import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import { initialFormState } from "../DataTable";
import FilterDialog from "../FilterDialog";

describe("FilterDialog", () => {
  const setActiveFilters = jest.fn();
  const filterList = {
    Name: ["app", "app2", "app3"],
    Status: ["Ready", "Failed"],
    Type: ["Application", "Helm Release"],
  };
  it("should not render when closed", () => {
    render(
      withTheme(
        <FilterDialog
          filterList={filterList}
          formState={initialFormState(filterList)}
          onFilterSelect={setActiveFilters}
          open={false}
        />
      )
    );
    expect(screen.getByTestId("container").getAttribute("class")).not.toContain(
      "open"
    );
  });
  it("should reveal filter list when open", () => {
    render(
      withTheme(
        <FilterDialog
          open
          formState={initialFormState(filterList)}
          filterList={filterList}
          onFilterSelect={setActiveFilters}
        />
      )
    );
    expect(screen.getByTestId("container").getAttribute("class")).toContain(
      "open"
    );
  });
  it("should return a value when a parameter is clicked", () => {
    const onFilterSelect = jest.fn();
    render(
      withTheme(
        <FilterDialog
          open
          formState={initialFormState(filterList)}
          filterList={filterList}
          onFilterSelect={onFilterSelect}
        />
      )
    );

    const checkbox1 = document.getElementById("Name:app") as HTMLInputElement;

    expect(checkbox1.checked).toEqual(false);
    fireEvent.click(checkbox1);

    const filterResult = onFilterSelect.mock.calls[0][0];
    expect(filterResult).toEqual({
      Name: ["app"],
    });
  });
});
