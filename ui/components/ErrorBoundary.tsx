import * as React from "react";
import Page from "./Page";

export default class ErrorBoundary extends React.Component<
  any,
  { hasError: boolean; error: Error }
> {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI.
    return { hasError: true, error };
  }

  componentDidCatch(error) {
    // You can also log the error to an error reporting service
    console.error(error);
  }

  render() {
    if (this.state.hasError) {
      // You can render any custom fallback UI
      return (
        <Page>
          <h3>Something went wrong.</h3>
          <pre>{this.state.error.message}</pre>
          <pre>{this.state.error.stack}</pre>
        </Page>
      );
    }

    return this.props.children;
  }
}
