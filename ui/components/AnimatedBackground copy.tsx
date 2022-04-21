import React from "react";

const OtherComponent = React.lazy(() => import("./AnimatedBackground"));

function MyComponent() {
  return (
    <React.Suspense fallback={null}>
      <OtherComponent />
    </React.Suspense>
  );
}

export default MyComponent;
