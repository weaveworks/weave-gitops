import * as React from "react";
import renderer from "react-test-renderer";
import { Core } from "../../lib/api/core/core.pb";
import CoreClientContextProvider, { CoreClientContext } from "../CoreClientContext"

describe("CoreContextProvider", () => {
  it("returns a non-empty api", () => {
    function TestComponent() {
      const { api } = React.useContext(CoreClientContext);
      expect(api.ListKustomizations).toBeTruthy();
      return (
        <div />
      );
    }

    renderer.create(
      <CoreClientContextProvider api={Core}><TestComponent /></CoreClientContextProvider>
    );

  });
});
