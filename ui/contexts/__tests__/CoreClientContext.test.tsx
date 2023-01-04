import * as React from "react";
import { QueryClient, QueryClientProvider } from "react-query";
import renderer from "react-test-renderer";
import { Core } from "../../lib/api/core/core.pb";
import CoreClientContextProvider, {
  CoreClientContext,
} from "../CoreClientContext";

describe("CoreContextProvider", () => {
  it("returns a non-empty api", () => {
    function TestComponent() {
      const { api } = React.useContext(CoreClientContext);
      expect(api.ListObjects).toBeTruthy();
      return <div />;
    }

    const queryClient = new QueryClient();

    renderer.create(
      <QueryClientProvider client={queryClient}>
        <CoreClientContextProvider api={Core}>
          <TestComponent />
        </CoreClientContextProvider>
      </QueryClientProvider>
    );
  });
});
