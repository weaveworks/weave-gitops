import React from "react";

const LottieAnimation = React.lazy(() => import("./LottieWrapper"));

function AnimatedBackground() {
  return (
    <React.Suspense fallback={null}>
      <LottieAnimation />
    </React.Suspense>
  );
}

export default AnimatedBackground;
