import { render, screen } from "@testing-library/react";
import * as React from "react";
import { withContext } from "../../lib/test-utils";
import useNavigation from "../navigation";

describe("useNavigation", () => {
  let container;
  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
  });
  afterEach(() => {
    document.body.removeChild(container);
    container = null;
  });

  it("returns the query", () => {
    const id = "custom-element";
    const myVar = "myVar";
    const TestComponent = () => {
      const { query } = useNavigation<{ someKey: string }>();

      return <p data-testid={id}>{query.someKey}</p>;
    };

    render(withContext(TestComponent, `/?someKey=${myVar}`));

    expect(screen.getByTestId(id).textContent).toEqual(myVar);
  });
});
