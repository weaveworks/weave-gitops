import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import * as React from "react";
import renderer from "react-test-renderer";
import { Core } from "../../lib/api/core/core.pb";
import CoreClientContextProvider, {
  CoreClientContext,
  CoreClientContextType,
} from "../CoreClientContext";

describe("CoreContextProvider", () => {
  it("returns a non-empty api", () => {
    function TestComponent() {
      const { api } = React.useContext(
        CoreClientContext,
      ) as CoreClientContextType;
      expect(api.ListObjects).toBeTruthy();
      return <div />;
    }

    const queryClient = new QueryClient();

    renderer.create(
      <QueryClientProvider client={queryClient}>
        <CoreClientContextProvider api={Core}>
          <TestComponent />
        </CoreClientContextProvider>
      </QueryClientProvider>,
    );
  });
});
