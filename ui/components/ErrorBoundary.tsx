import React, { useEffect, useState } from "react";
import { useLocation } from "react-router";
import Page from "./Page";

interface Props {
  hasError?: boolean;
  error?: Error | null;
  children?: any;
  setHasError?: (hasError: boolean) => void;
}

class ErrorBoundaryDetail extends React.Component<Props, any> {
  constructor(props: any) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidUpdate(prevProps: Props) {
    if (!this.props.hasError && prevProps.hasError) {
      this.setState({ hasError: false });
    }
  }

  componentDidCatch(error: Error) {
    console.error(error);
    this.props.setHasError(true);
  }

  render() {
    if (this.state.hasError) {
      return (
        <Page path={[]}>
          <h3>Something went wrong.</h3>
          <pre>{this.state.error?.message}</pre>
          <pre>{this.state.error?.stack}</pre>
        </Page>
      );
    }

    return this.props.children;
  }
}

/** Function component wrapper as we need useEffect to set the state back to false on location changing **/
function ErrorBoundary({ children }: Props) {
  const [hasError, setHasError] = useState<boolean>(false);
  const location = useLocation();

  useEffect(() => {
    if (hasError) {
      setHasError(false);
    }
  }, [location.key]);

  return (
    <ErrorBoundaryDetail hasError={hasError} setHasError={setHasError}>
      {children}
    </ErrorBoundaryDetail>
  );
}

export default ErrorBoundary;
