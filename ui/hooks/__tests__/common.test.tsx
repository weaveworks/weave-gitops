import { render, screen } from "@testing-library/react";
import * as React from "react";
import { act } from "react-dom/test-utils";
import { useRequestState } from "../common";

describe("useRequestState", () => {
  let container;
  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
  });
  afterEach(() => {
    document.body.removeChild(container);
    container = null;
  });
  it("shows loading", async () => {
    const id = "loading";
    const text = "loading...";
    type res = { foo: string };
    const TestComponent = () => {
      const [, loading, , req] = useRequestState<res>();

      React.useEffect(() => {
        const p = new Promise<res>(() => {
          //   Never continue
        });
        req(p);
      }, []);

      return <div data-testid={id}>{loading && text}</div>;
    };

    await act(async () => {
      render(<TestComponent />);
    });

    expect(await (await screen.findByTestId(id)).textContent).toEqual(text);
  });
  it("shows an error", async () => {
    const id = "error";
    const errorMsg = "There was a massive problem";
    type res = { foo: string };
    const TestComponent = () => {
      const [, , error, req] = useRequestState<res>();

      React.useEffect(() => {
        const p = new Promise<res>((accept, reject) => {
          reject(errorMsg);
        });
        req(p);
      }, []);

      if (error) {
        return <div data-testid={id}>{error}</div>;
      }
      return <div>all good</div>;
    };

    await act(async () => {
      render(<TestComponent />);
    });

    expect(await (await screen.findByTestId(id)).textContent).toEqual(errorMsg);
  });
  it("shows the resolved value", async () => {
    const id = "value";
    const foo = "bar";
    type res = { foo: string };
    const TestComponent = () => {
      const [value, loading, , req] = useRequestState<res>();

      React.useEffect(() => {
        const p = new Promise<res>((accept) => {
          accept({ foo });
        });
        req(p);
      }, []);

      if (loading) {
        return <div>loading...</div>;
      }

      return <div data-testid={id}>{value.foo}</div>;
    };

    await act(async () => {
      render(<TestComponent />);
    });

    expect(await (await screen.findByTestId(id)).textContent).toEqual(foo);
  });
});
